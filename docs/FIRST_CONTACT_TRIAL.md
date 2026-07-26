# First Contact Trial

This trial is for a person who did not help build Logos-Formal and has not been
briefed on its private terminology.

## First Contact Evidence Log

| # | Date (UTC) | Actor | Trial version / hash | Elapsed | What they said it does (short) | Friction? | Link |
|---|------------|-------|----------------------|---------|--------------------------------|-----------|------|
| 1 | — | — | — | — | — | — | — |
| 2 | — | — | — | — | — | — | — |
| 3 | — | — | — | — | — | — | — |

**Contract status**

- Qualifying participants: **0 / 3**
- Verified friction events: **0 / 1**
- Overall: **`EVIDENCE_INCOMPLETE`**

**Rules**

- Only independent actors count.
- Internal CI and Principal runs do not count.
- Friction must be reproducible against the published surface.
- Duplicate identities—one person using different handles—count as one
  participant unless they are shown to be distinct actors.

## Qualifying participant

A report counts toward the First Contact threshold only when all of these
conditions are satisfied:

1. **Independence**
   - The participant is not the project Principal.
   - The participant is not an agent acting under the Principal's direct
     session control for the purpose of producing evidence.
   - The participant has no write access to protected `main` or to the
     evidence log.
2. **Actual run**
   - The participant uses the published portable trial or the exact
     PR-candidate cold-run path below.
   - The run uses the declared source commit or package hash.
   - The run occurs on the participant's own machine or environment.
3. **Minimum report content**
   - elapsed time from start to a working result or clear failure;
   - one short statement of what the participant believes the surface does;
   - friction, confusion, or blockage, or an explicit statement that none was
     observed.
4. **Verifiability**
   - An Issue #2 comment needs no additional identity machinery: the GitHub
     handle and comment timestamp are sufficient.
   - A report from another public channel includes a link with visible
     identity and timestamp.
   - A private report is reposted by the participant to Issue #2, or the
     maintainer records a dated, redacted summary identifying it as an
     independent report.

Stars, forks, praise without a run, internal CI, project-authored runs,
agent-generated evidence rehearsals, locally modified surfaces, and reports
missing either elapsed time or the participant's explanation do not count.

## Instructions for the participant

The preferred route is the portable package:

```text
dist/cyonic-node-setup-v0.1.2-portable-7e0f263.zip
sha256:8d1263b1dff0fcd84c42ac8911f132d270e8b172d50d06982d73d227499315a7
```

After verifying and unpacking the archive, run:

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

The source-checkout alternative follows.

Start timing before cloning:

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

The service is currently a public pull-request candidate, not part of default
`main`. Record the exact commit printed by `git rev-parse HEAD`; do not replace
it with a branch name in the report. The published service candidate baseline
is:

```text
7e0f263c2a31eaece5bf8ded76ce4ba47c7db32a
```

Answer from the command output rather than from other project documents.

The report path is create-only. Use a new filename for every attempt; the
command will refuse to replace an existing participant report.

The report records:

- the exact Git source ref when available;
- the boundary receipt;
- self-reported setup time;
- the participant's own explanation;
- the participant's friction category and description;
- `CLAIMED_EXTERNAL`;
- `PENDING_EXTERNAL_REVIEW`;
- `authorityEffect: NONE`.

## Minimum project record

Do not edit the participant's original report.

For each qualifying report, fill exactly one row in the First Contact Evidence
Log above. Keep the participant's short explanation in their own words.

Before counting the line, check:

1. the participant satisfies the independence conditions above;
2. the source ref or archive hash matches the published trial;
3. the report has a durable identity/time record;
4. any generated receipt still shows `010`, `OBSERVE_ONLY`,
   `forwarded: false`, and `NOT_PERFORMED`;
5. the participant's explanation distinguishes:
   - structural interpretation;
   - externally supplied authorization evidence;
   - absence of effect;
6. elapsed time is present;
7. any claimed friction is a reproducible limitation or confusion affecting
   use of the published surface.

For Phase 1, an Issue #2 comment plus this one-line record is sufficient. Do
not require cryptographic timestamps, stronger identity proof, or automated
evidence ingestion.

## Optional structured adjudication

The existing CLI can bind a later structured review to the exact report bytes.
It is optional for the initial 0-to-3 evidence gate and should not replace the
simple public record above.

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

PowerShell:

```powershell
go run ./runtime/cyonic-service adjudicate `
  -report cyonic-first-contact-report-01.json `
  -out cyonic-first-contact-adjudication-01.json `
  -reviewer independent-reviewer-01 `
  -externality VERIFIED_EXTERNAL `
  -source VERIFIED `
  -comprehension ACCEPTED `
  -friction ACCEPTED `
  -notes "Participant relationship and source ref checked independently."
```

When used, the command does not edit the report. It records:

- `sha256` of the exact report bytes;
- reviewer findings as explicit attestations;
- a mechanical check of the embedded 010/no-effect receipt shape;
- a mechanical check of the self-reported ten-minute threshold;
- `QUALIFYING_EXTERNAL` only when all required findings pass;
- `VERIFIED_EXTERNAL_FRICTION` only for a qualifying report with concrete,
  reviewer-accepted friction;
- `authorityEffect: NONE`.

Evidence outputs are create-only. The CLI refuses to overwrite participant
reports, rejects an adjudication output path that equals its participant
report, and refuses to overwrite any existing adjudication or summary file.

If structured adjudications are later collected, they can be summarized:

```bash
go run ./runtime/cyonic-service summarize-trials \
  -dir evidence/adjudications \
  -out cyonic-first-contact-summary-01.json
```

The automated summary is not required for Phase 1. Its threshold remains three
qualifying participants and at least one verified external friction event.
That threshold is not adoption, certification, canon elevation, or authority.

## Evidence limits

One report is one interaction, not adoption. Three qualifying cold reports are
required for `FIRST_CONTACT_VALIDATED`.

An external friction event requires both:

- a qualifying external participant;
- a concrete limitation or confusion affecting use of the published surface;
- independent reproduction or verification of that limitation or confusion.

A preference or feature request alone is not friction. A friction record is
evidence input and does not authorize a design change.

Internal runs, project-authored answers, model-generated rehearsals, runs
against unpublished or locally modified surfaces, and incomplete reports do
not count.

Reviewer findings are attestations, not self-proving facts. Until a signature
profile is separately specified, retain the reviewer record and its delivery
context with the adjudication file.

Do not add evidence-collection automation before at least one real external
report exists. Issue #2 comments and the minimum project record are sufficient
for the current gate.

## Current evidence state

Internal rehearsals and CI runs exist, but they do not satisfy this trial.
Until participant reports and additive adjudications are present:

```text
qualifying external participants: 0/3
verified external friction events: 0/1
First Contact status: EVIDENCE_INCOMPLETE
```
