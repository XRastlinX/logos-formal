import { DatabaseSync } from 'node:sqlite';

export class CadenceEngine {
  private db: DatabaseSync;
  private isRunning = false;
  private tickIntervalMs: number;
  private timer: NodeJS.Timeout | null = null;
  private ttlSeconds: number;

  constructor(db: DatabaseSync, tickIntervalMs = 500, ttlSeconds = 30) {
    this.db = db;
    this.tickIntervalMs = tickIntervalMs;
    this.ttlSeconds = ttlSeconds;
  }

  public start(): void {
    if (this.isRunning) return;
    this.isRunning = true;
    this.tick();
  }

  public stop(): void {
    this.isRunning = false;
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }
  }

  private tick(): void {
    if (!this.isRunning) return;

    try {
      this.enforceTimeouts();
    } catch (err) {
      console.error('Cadence sweep failed, recovering next tick:', err);
    } finally {
      // Schedule the next tick only after the current one completes
      if (this.isRunning) {
        this.timer = setTimeout(() => this.tick(), this.tickIntervalMs);
      }
    }
  }

  public enforceTimeouts(asOf = new Date()): void {
    // Phase 1: Read-only sweep (Non-blocking)
    // Using unixepoch to calculate difference since sqlite CURRENT_TIMESTAMP / initiated_at is an ISO string.
    // Or we can use (julianday('now') - julianday(initiated_at)) * 86400
    const sweep = this.db.prepare(`
      SELECT obligation_id
      FROM matrix_obligations
      WHERE current_state = 'SUSPENDED_STATE'
        AND op_cost = 'INFINITY'
        AND (julianday(?) - julianday(initiated_at)) * 86400 > ?
    `);

    const stalledObligations = sweep.all(
      asOf.toISOString(),
      this.ttlSeconds,
    ) as { obligation_id: string }[];
    if (stalledObligations.length === 0) return;

    // Phase 2: The Strike (Optimistic Concurrency Lock)
    const strike = this.db.prepare(`
      UPDATE matrix_obligations
      SET current_state = 'TERMINAL_TIMEOUT',
          op_cost = '0'
      WHERE obligation_id = ?
        AND current_state = 'SUSPENDED_STATE'
    `);

    try {
      this.db.exec('BEGIN IMMEDIATE;');
      let resolvedCount = 0;
      for (const obs of stalledObligations) {
        const info = strike.run(obs.obligation_id);
        resolvedCount += Number(info.changes);
        // info.changes will be 0 if a receipt cleared it milliseconds prior
      }
      this.db.exec('COMMIT;');
    } catch (e) {
      try {
        this.db.exec('ROLLBACK;');
      } catch {
        // Ignore rollback failure
      }
      throw e;
    }
  }
}
