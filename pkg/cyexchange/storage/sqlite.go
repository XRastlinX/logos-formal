// Status: PROPOSED
// authority_effect: NONE
package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	sqlite "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
	cyerrors "github.com/XRastlinX/logos-formal/pkg/cyexchange/errors"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/inbox"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/outbox"
)

const (
	queueInbox  = "INBOX"
	queueOutbox = "OUTBOX"
)

type SQLiteStore struct {
	db    *sql.DB
	clock Clock
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	return NewSQLiteStoreWithClock(dbPath, time.Now)
}

func NewSQLiteStoreWithClock(dbPath string, clock Clock) (*SQLiteStore, error) {
	if dbPath == "" || clock == nil {
		return nil, fmt.Errorf("%w: database path and clock are required", cyerrors.ErrValidation)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("%w: open database: %v", cyerrors.ErrDurability, err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	store := &SQLiteStore{db: db, clock: clock}
	if err := store.initialize(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *SQLiteStore) initialize() error {
	statements := []string{
		`PRAGMA foreign_keys = ON`,
		`PRAGMA journal_mode = WAL`,
		`PRAGMA synchronous = FULL`,
		`CREATE TABLE IF NOT EXISTS queue_messages (
			queue_kind TEXT NOT NULL CHECK (queue_kind IN ('INBOX', 'OUTBOX')),
			envelope_id TEXT NOT NULL,
			envelope_json BLOB NOT NULL,
			expires_at_ns INTEGER NOT NULL,
			created_at_ns INTEGER NOT NULL,
			PRIMARY KEY (queue_kind, envelope_id)
		) WITHOUT ROWID`,
		`CREATE TABLE IF NOT EXISTS queue_events (
			event_id INTEGER PRIMARY KEY AUTOINCREMENT,
			queue_kind TEXT NOT NULL,
			envelope_id TEXT NOT NULL,
			state INTEGER NOT NULL,
			recorded_at_ns INTEGER NOT NULL,
			FOREIGN KEY (queue_kind, envelope_id)
				REFERENCES queue_messages(queue_kind, envelope_id)
		)`,
		`CREATE INDEX IF NOT EXISTS queue_events_lookup
			ON queue_events(queue_kind, envelope_id, event_id)`,
	}
	for _, statement := range statements {
		if _, err := s.db.Exec(statement); err != nil {
			return fmt.Errorf("%w: initialize database: %v", cyerrors.ErrDurability, err)
		}
	}
	return nil
}

func (s *SQLiteStore) InsertEnvelope(
	env *envelope.CyExchangeEnvelope,
	expiresAt time.Time,
) error {
	if env == nil || env.EnvelopeID == "" || expiresAt.IsZero() {
		return fmt.Errorf("%w: invalid inbox insert", cyerrors.ErrValidation)
	}
	return s.insert(
		queueInbox,
		env,
		expiresAt.UTC().UnixNano(),
		int(inbox.StateReceived),
	)
}

func (s *SQLiteStore) EnqueueOutbox(
	env *envelope.CyExchangeEnvelope,
	queuedAt time.Time,
) error {
	if env == nil || env.EnvelopeID == "" || queuedAt.IsZero() {
		return fmt.Errorf("%w: invalid outbox insert", cyerrors.ErrValidation)
	}
	return s.insert(queueOutbox, env, 0, int(outbox.StateQueued))
}

func (s *SQLiteStore) insert(
	queueKind string,
	env *envelope.CyExchangeEnvelope,
	expiresAtNS int64,
	initialState int,
) error {
	payload, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("%w: marshal envelope: %v", cyerrors.ErrValidation, err)
	}
	nowNS := s.clock().UTC().UnixNano()
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%w: begin insert: %v", cyerrors.ErrDurability, err)
	}
	defer tx.Rollback()
	_, err = tx.Exec(
		`INSERT INTO queue_messages
			(queue_kind, envelope_id, envelope_json, expires_at_ns, created_at_ns)
		 VALUES (?, ?, ?, ?, ?)`,
		queueKind,
		env.EnvelopeID,
		payload,
		expiresAtNS,
		nowNS,
	)
	if err != nil {
		if isConstraint(err) {
			return fmt.Errorf("%w: duplicate %s envelope %s", cyerrors.ErrReplay, queueKind, env.EnvelopeID)
		}
		return fmt.Errorf("%w: insert envelope: %v", cyerrors.ErrDurability, err)
	}
	if _, err := tx.Exec(
		`INSERT INTO queue_events
			(queue_kind, envelope_id, state, recorded_at_ns)
		 VALUES (?, ?, ?, ?)`,
		queueKind,
		env.EnvelopeID,
		initialState,
		nowNS,
	); err != nil {
		return fmt.Errorf("%w: insert initial event: %v", cyerrors.ErrDurability, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: commit insert: %v", cyerrors.ErrDurability, err)
	}
	return nil
}

func (s *SQLiteStore) UpdateState(envelopeID string, newState inbox.State) error {
	return s.appendTransition(
		queueInbox,
		envelopeID,
		int(newState),
		func(current int) bool {
			return inbox.CanTransition(inbox.State(current), newState)
		},
	)
}

func (s *SQLiteStore) UpdateOutboxState(
	envelopeID string,
	newState outbox.State,
) error {
	return s.appendTransition(
		queueOutbox,
		envelopeID,
		int(newState),
		func(current int) bool {
			return outbox.CanTransition(outbox.State(current), newState)
		},
	)
}

func (s *SQLiteStore) appendTransition(
	queueKind, envelopeID string,
	newState int,
	allowed func(current int) bool,
) error {
	if envelopeID == "" || allowed == nil {
		return fmt.Errorf("%w: invalid transition request", cyerrors.ErrValidation)
	}
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%w: begin transition: %v", cyerrors.ErrDurability, err)
	}
	defer tx.Rollback()
	var current int
	err = tx.QueryRow(
		`SELECT state FROM queue_events
		  WHERE queue_kind = ? AND envelope_id = ?
		  ORDER BY event_id DESC LIMIT 1`,
		queueKind,
		envelopeID,
	).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return cyerrors.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("%w: read current state: %v", cyerrors.ErrDurability, err)
	}
	if !allowed(current) {
		return fmt.Errorf(
			"%w: illegal %s transition %d -> %d",
			cyerrors.ErrValidation,
			queueKind,
			current,
			newState,
		)
	}
	if _, err := tx.Exec(
		`INSERT INTO queue_events
			(queue_kind, envelope_id, state, recorded_at_ns)
		 VALUES (?, ?, ?, ?)`,
		queueKind,
		envelopeID,
		newState,
		s.clock().UTC().UnixNano(),
	); err != nil {
		return fmt.Errorf("%w: append transition: %v", cyerrors.ErrDurability, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: commit transition: %v", cyerrors.ErrDurability, err)
	}
	return nil
}

func (s *SQLiteStore) GetEnvelope(
	envelopeID string,
) (*envelope.CyExchangeEnvelope, inbox.State, error) {
	raw, state, _, _, err := s.get(queueInbox, envelopeID)
	if err != nil {
		return nil, inbox.StateUnknown, err
	}
	env, err := decodeStoredEnvelope(raw)
	if err != nil {
		return nil, inbox.StateUnknown, err
	}
	return env, inbox.State(state), nil
}

func (s *SQLiteStore) GetOutbox(
	envelopeID string,
) (*envelope.CyExchangeEnvelope, outbox.State, error) {
	raw, state, _, _, err := s.get(queueOutbox, envelopeID)
	if err != nil {
		return nil, outbox.StateUnknown, err
	}
	env, err := decodeStoredEnvelope(raw)
	if err != nil {
		return nil, outbox.StateUnknown, err
	}
	return env, outbox.State(state), nil
}

func (s *SQLiteStore) get(
	queueKind, envelopeID string,
) ([]byte, int, int64, int64, error) {
	var raw []byte
	var state int
	var expiresAtNS, createdAtNS int64
	err := s.db.QueryRow(
		`SELECT m.envelope_json, m.expires_at_ns, m.created_at_ns,
		        (SELECT e.state FROM queue_events e
		          WHERE e.queue_kind = m.queue_kind
		            AND e.envelope_id = m.envelope_id
		          ORDER BY e.event_id DESC LIMIT 1)
		   FROM queue_messages m
		  WHERE m.queue_kind = ? AND m.envelope_id = ?`,
		queueKind,
		envelopeID,
	).Scan(&raw, &expiresAtNS, &createdAtNS, &state)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, 0, 0, cyerrors.ErrNotFound
	}
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("%w: get queue record: %v", cyerrors.ErrDurability, err)
	}
	return raw, state, expiresAtNS, createdAtNS, nil
}

func (s *SQLiteStore) GetRecoverableLog() ([]*inbox.Record, error) {
	rows, err := s.loadQueue(queueInbox)
	if err != nil {
		return nil, err
	}
	records := make([]*inbox.Record, 0, len(rows))
	for _, row := range rows {
		states, err := s.loadStates(queueInbox, row.envelopeID)
		if err != nil {
			return nil, err
		}
		if err := validateInboxHistory(states); err != nil {
			return nil, fmt.Errorf("%w: envelope %s: %v", cyerrors.ErrCorrupt, row.envelopeID, err)
		}
		env, err := decodeStoredEnvelope(row.rawEnvelope)
		if err != nil {
			return nil, err
		}
		records = append(records, &inbox.Record{
			Envelope:  env,
			State:     inbox.State(states[len(states)-1]),
			ExpiresAt: time.Unix(0, row.expiresAtNS).UTC(),
		})
	}
	sortInboxRecords(records)
	return records, nil
}

func (s *SQLiteStore) GetRecoverableOutbox() ([]*outbox.Record, error) {
	rows, err := s.loadQueue(queueOutbox)
	if err != nil {
		return nil, err
	}
	records := make([]*outbox.Record, 0, len(rows))
	for _, row := range rows {
		states, err := s.loadStates(queueOutbox, row.envelopeID)
		if err != nil {
			return nil, err
		}
		if err := validateOutboxHistory(states); err != nil {
			return nil, fmt.Errorf("%w: envelope %s: %v", cyerrors.ErrCorrupt, row.envelopeID, err)
		}
		env, err := decodeStoredEnvelope(row.rawEnvelope)
		if err != nil {
			return nil, err
		}
		records = append(records, &outbox.Record{
			Envelope: env,
			State:    outbox.State(states[len(states)-1]),
			QueuedAt: time.Unix(0, row.createdAtNS).UTC(),
		})
	}
	sortOutboxRecords(records)
	return records, nil
}

type queueRow struct {
	envelopeID  string
	rawEnvelope []byte
	expiresAtNS int64
	createdAtNS int64
}

func (s *SQLiteStore) loadQueue(queueKind string) ([]queueRow, error) {
	rows, err := s.db.Query(
		`SELECT envelope_id, envelope_json, expires_at_ns, created_at_ns
		   FROM queue_messages WHERE queue_kind = ? ORDER BY envelope_id`,
		queueKind,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: load queue: %v", cyerrors.ErrDurability, err)
	}
	defer rows.Close()
	var result []queueRow
	for rows.Next() {
		var row queueRow
		if err := rows.Scan(
			&row.envelopeID,
			&row.rawEnvelope,
			&row.expiresAtNS,
			&row.createdAtNS,
		); err != nil {
			return nil, fmt.Errorf("%w: scan queue: %v", cyerrors.ErrDurability, err)
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: iterate queue: %v", cyerrors.ErrDurability, err)
	}
	return result, nil
}

func (s *SQLiteStore) loadStates(queueKind, envelopeID string) ([]int, error) {
	rows, err := s.db.Query(
		`SELECT state FROM queue_events
		  WHERE queue_kind = ? AND envelope_id = ?
		  ORDER BY event_id`,
		queueKind,
		envelopeID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w: load state history: %v", cyerrors.ErrDurability, err)
	}
	defer rows.Close()
	var states []int
	for rows.Next() {
		var state int
		if err := rows.Scan(&state); err != nil {
			return nil, fmt.Errorf("%w: scan state history: %v", cyerrors.ErrDurability, err)
		}
		states = append(states, state)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: iterate state history: %v", cyerrors.ErrDurability, err)
	}
	if len(states) == 0 {
		return nil, fmt.Errorf("%w: missing state history", cyerrors.ErrCorrupt)
	}
	return states, nil
}

func (s *SQLiteStore) EventCount(queueKind, envelopeID string) (int, error) {
	if queueKind != queueInbox && queueKind != queueOutbox {
		return 0, fmt.Errorf("%w: invalid queue kind", cyerrors.ErrValidation)
	}
	var count int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM queue_events
		  WHERE queue_kind = ? AND envelope_id = ?`,
		queueKind,
		envelopeID,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf("%w: count events: %v", cyerrors.ErrDurability, err)
	}
	if count == 0 {
		return 0, cyerrors.ErrNotFound
	}
	return count, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func decodeStoredEnvelope(raw []byte) (*envelope.CyExchangeEnvelope, error) {
	var env envelope.CyExchangeEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, fmt.Errorf("%w: decode stored envelope: %v", cyerrors.ErrCorrupt, err)
	}
	return &env, nil
}

func validateInboxHistory(states []int) error {
	if len(states) == 0 || inbox.State(states[0]) != inbox.StateReceived {
		return errors.New("inbox history does not begin at RECEIVED")
	}
	for i := 1; i < len(states); i++ {
		if !inbox.CanTransition(inbox.State(states[i-1]), inbox.State(states[i])) {
			return fmt.Errorf("illegal inbox history %d -> %d", states[i-1], states[i])
		}
	}
	last := inbox.State(states[len(states)-1])
	if last == inbox.StateEffectRequested || last == inbox.StateApplied {
		return errors.New("effect state is forbidden in 010 store")
	}
	return nil
}

func validateOutboxHistory(states []int) error {
	if len(states) == 0 || outbox.State(states[0]) != outbox.StateQueued {
		return errors.New("outbox history does not begin at QUEUED")
	}
	for i := 1; i < len(states); i++ {
		if !outbox.CanTransition(outbox.State(states[i-1]), outbox.State(states[i])) {
			return fmt.Errorf("illegal outbox history %d -> %d", states[i-1], states[i])
		}
	}
	return nil
}

func isConstraint(err error) bool {
	var sqliteError *sqlite.Error
	if !errors.As(err, &sqliteError) {
		return false
	}
	code := sqliteError.Code()
	return code == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY ||
		code == sqlite3.SQLITE_CONSTRAINT_UNIQUE
}
