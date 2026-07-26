# First Contact Trial

This trial is for a person who did not help build Logos-Formal and has not been
briefed on its private terminology.

## Instructions for the participant

Start timing before cloning:

```bash
git clone https://github.com/XRastlinX/logos-formal.git
cd logos-formal
go run ./runtime/cyonic-service trial \
  -origin external \
  -report cyonic-first-contact-report-01.json
```

PowerShell:

```powershell
git clone https://github.com/XRastlinX/logos-formal.git
Set-Location logos-formal
go run ./runtime/cyonic-service trial `
  -origin external `
  -report cyonic-first-contact-report-01.json
```

Answer from the command output rather than from other project documents.

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

## Evidence limits

One report is one interaction, not adoption. Three qualifying cold reports are
required for `FIRST_CONTACT_VALIDATED`.

An external friction event requires both:

- a qualifying external participant;
- a concrete friction description that an independent reviewer accepts.

Internal runs, project-authored answers, and model-generated rehearsals do not
count.
