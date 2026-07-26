# Contributing to Logos-Formal

Thank you for your interest in contributing to Logos-Formal.

## How to Contribute

### Report Issues

If you find a bug, a broken invariant, or a gap in the formal boundaries, [open an issue](https://github.com/XRastlinX/logos-formal/issues). Include:

- What you expected to happen
- What actually happened
- Steps to reproduce (if applicable)
- Which demonstrator or document is affected

### Suggest Improvements

Open an issue tagged `enhancement` for:

- New negative test vectors for the demonstrators
- Improvements to the formal specifications
- Standards alignment corrections (RATS, ABAC, in-toto, SLSA)
- Documentation clarity improvements

### Submit Code

1. Fork the repository
2. Create a branch from `main`
3. Make your changes
4. Ensure `go test ./...` passes
5. Ensure `go vet ./...` passes
6. Submit a pull request

### Epistemic Rules

All contributions must maintain the project's epistemic discipline:

- **Do not inflate claims.** Use "demonstrates," not "proves."
- **Do not elevate PROPOSED artifacts.** The Autophagic Operator (α) is research, not canon.
- **Preserve provenance.** If you reference external work, cite it.
- **Maintain fail-closed semantics.** Negative test vectors must cause actual test failures, not just print error messages.
- **Authority Effect = NONE.** No contribution grants operational authority.

### Code Style

- Go code follows standard `gofmt` formatting
- Markdown follows GitHub Flavored Markdown
- Commit messages are descriptive and reference the affected domain

## Code of Conduct

Be precise, be honest, be constructive. Disagree on substance, not on style.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
