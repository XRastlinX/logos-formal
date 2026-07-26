# First Contact Trial

This trial is for a person who did not help build Logos-Formal and has not been
briefed on its private terminology.

## Instructions for the participant

Start timing before cloning:

```bash
git clone \
  --branch codex/cyonic-service-surface-v1 \
  --single-branch \
  https://github.com/XRastlinX/logos-formal.git
cd logos-formal
git rev-parse HEAD
go run ./runtime/cyonic-service trial \
  -origin external \
  -report cyonic-first-contact-report-01.json
```

PowerShell:

```powershell
git clone `
  --branch codex/cyonic-service-surface-v1 `
  --single-branch `
  https://github.com/XRastlinX/logos-formal.git
Set-Location logos-formal
git rev-parse HEAD
go run ./runtime/cyonic-service trial `
  -origin external `
  -report cyonic-first-contact-report-01.json
```

The service is currently a public pull-request candidate, not part of default
`main`. Record the exact commit printed by `git rev-parse HEAD`; do not replace
it with a branch name in the report. The current independently tested candidate
baseline is:

```text
3a5b0801bc243a1716273dfe98acc5265e6c6882
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

## Instructions for the project reviewer

Do not edit the participant's original report.

Verify separately:

1. the participant was not part of the project team;
2. the source ref exists and contains the trial implementation;
3. the report receipt still shows `010`, `OBSERVE_ONLY`,
   `forwarded: false`, and `NOT_PERFORMED`;
4. the participant's explanation distinguishes:
   - structural interpretation;
   - externally supplied authorization evidence;
   - absence of effect;
5. setup time is below ten minutes;
6. any claimed friction is reproducible or sufficiently described.

Create an additive adjudication record. Never replace
`PENDING_EXTERNAL_REVIEW` inside the participant's report.

The included CLI binds the review to the exact report bytes:

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

The command does not edit the report. It records:

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

When multiple additive adjudications have been collected:

```bash
go run ./runtime/cyonic-service summarize-trials \
  -dir evidence/adjudications \
  -out cyonic-first-contact-summary-01.json
```

The summary deduplicates qualifying participants by `participantRef` and
friction events by report digest. It reports `FIRST_CONTACT_VALIDATED` only
when there are at least three qualifying participants and at least one
verified external friction event. This status is a Phase 1 evidence threshold,
not adoption, certification, canon elevation, or authority.

## Evidence limits

One report is one interaction, not adoption. Three qualifying cold reports are
required for `FIRST_CONTACT_VALIDATED`.

An external friction event requires both:

- a qualifying external participant;
- a concrete friction description that an independent reviewer accepts.

Internal runs, project-authored answers, and model-generated rehearsals do not
count.

Reviewer findings are attestations, not self-proving facts. Until a signature
profile is separately specified, retain the reviewer record and its delivery
context with the adjudication file.

## Current evidence state

Internal rehearsals and CI runs exist, but they do not satisfy this trial.
Until participant reports and additive adjudications are present:

```text
qualifying external participants: 0/3
verified external friction events: 0/1
First Contact status: EVIDENCE_INCOMPLETE
```
