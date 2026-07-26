# @# D08 — Security Policy

## Report privately

Do not place credentials, private keys, access tokens, cookies, private packet
contents, or exploitable vulnerabilities in a public issue.

When the repository is hosted on GitHub, use the repository's private security
advisory channel or contact the repository owner through an established
private channel.

## Protected surfaces

- packet path and archive validation;
- content-addressed manifests and receipts;
- connector scopes and leases;
- quarantine boundaries;
- CI workflows;
- CODEOWNERS and review rules.

## Non-authority rule

A security scan, signature, or passing check supplies evidence. It does not
create Owner, Seal, Apply, or state-commit authority.
