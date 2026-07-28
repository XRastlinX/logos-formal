# Optional First Contact Trial

This trial is an optional way for someone who did not build Logos-Formal to
report cold-run usability. It is not a merge gate, release gate, adoption
claim, certification process, or source of authority.

```text
external evaluation: NOT ESTABLISHED
engineering verification: INTERNAL TESTS PASS
authority effect: NONE
```

Independent feedback is useful but not required for engineering work,
publication, merge, or continued development. Friction is recorded when
observed; no friction event is required.

## Feedback log

| Date (UTC) | Actor | Trial version / hash | Elapsed | What they said it does | Friction? | Link |
|---|---|---|---|---|---|---|
| — | — | — | — | — | — | — |

## What makes a report externally informative

An externally informative report normally includes:

1. a participant who did not build the service;
2. the published package hash or exact source commit;
3. a run on the participant's own machine or environment;
4. elapsed time;
5. the participant's explanation of the boundary;
6. concrete friction, or an explicit statement that none was observed;
7. a durable link or recorded delivery context.

Maintainer runs, CI, and internal rehearsals remain valid engineering evidence
when labeled internal. They are simply not external evaluation.

## Preferred path: portable package

```text
dist/cyonic-node-setup-v0.1.2-portable-7e0f263.zip
sha256:8d1263b1dff0fcd84c42ac8911f132d270e8b172d50d06982d73d227499315a7
```

After verifying and unpacking the archive, run on Windows:

```powershell
.\setup.cmd friend-node-01 codex
.\trial.cmd cyonic-first-contact-report-01.json
```

Or on Linux/macOS:

```bash
chmod +x scripts/*.sh
./scripts/setup.sh friend-node-01 codex
./scripts/trial.sh cyonic-first-contact-report-01.json
```

## Alternative path: exact source checkout

Bash:

```bash
expected_commit="7e0f263c2a31eaece5bf8ded76ce4ba47c7db32a"
git clone https://github.com/XRastlinX/logos-formal.git
cd logos-formal
git checkout --detach "$expected_commit"
test "$(git rev-parse HEAD)" = "$expected_commit" || exit 1
go run ./runtime/cyonic-service trial \
  -origin external \
  -report cyonic-first-contact-report-01.json
```

PowerShell:

```powershell
$ExpectedCommit = "7e0f263c2a31eaece5bf8ded76ce4ba47c7db32a"
git clone https://github.com/XRastlinX/logos-formal.git
Set-Location logos-formal
git checkout --detach $ExpectedCommit
$ActualCommit = (git rev-parse HEAD).Trim()
if ($ActualCommit -ne $ExpectedCommit) {
    throw "Exact candidate verification failed; do not continue."
}
go run ./runtime/cyonic-service trial `
  -origin external `
  -report cyonic-first-contact-report-01.json
```

The report path is create-only. Use a new filename for each attempt.

The report records:

- the exact Git source ref when available;
- the boundary receipt;
- self-reported setup time;
- the participant's own explanation;
- any friction category and description;
- `CLAIMED_EXTERNAL`;
- `PENDING_EXTERNAL_REVIEW`;
- `authorityEffect: NONE`.

## Reporting

Issue #2 is the preferred public feedback channel. A report may include:

1. operating system and `go version`;
2. exact commit or package hash;
3. elapsed time;
4. the participant's explanation of interpretation, external authorization
   evidence, and effect;
5. terminal output showing a failure, if any;
6. concrete friction or `NONE`.

Do not post credentials, private keys, access tokens, or unrelated personal
paths.

## Optional structured adjudication

The CLI can bind an optional review to exact report bytes:

```bash
go run ./runtime/cyonic-service adjudicate \
  -report cyonic-first-contact-report-01.json \
  -out cyonic-first-contact-adjudication-01.json \
  -reviewer independent-reviewer-01 \
  -externality VERIFIED_EXTERNAL \
  -source VERIFIED \
  -comprehension ACCEPTED \
  -friction ACCEPTED \
  -notes "Participant relationship and source ref checked independently."
```

Structured adjudication is additive and create-only. It does not edit the
participant report or create authority.

Collected adjudications may be summarized:

```bash
go run ./runtime/cyonic-service summarize-trials \
  -dir evidence/adjudications \
  -out cyonic-first-contact-summary-01.json
```

The summary records deduplicated reports and observed friction. It has no
completion threshold and does not govern merge eligibility.

## Evidence limits

- One report is one interaction, not adoption.
- Three reports are still only three interactions.
- Friction is a reported limitation, not a required outcome.
- Absence of friction is a valid observation.
- Reviewer findings are attestations, not self-proving facts.
- No report count authorizes a design change, effect, or canon elevation.
- Internal verification supports only the behavior covered by its declared
  tests.

Current external status:

```text
external evaluation: NOT ESTABLISHED
external reports recorded: 0
verified external friction recorded: 0
authority effect: NONE
```
