# Repository Guidelines

## Project Structure & Module Organization
Core client logic sits at the module root alongside feature-specific files such as `message.go`, `presence.go`, and `upload.go`. Shared protocol definitions are under `proto/`, transport plumbing in `socket/`, and storage abstractions in `store/` (with `store/sqlstore` providing SQL-backed persistence). Cryptographic helpers and logging utilities live in `util/`, while device state fixtures reside in `appstate/`. Tests belong next to the code they cover; see `client_test.go` for layout and package conventions.

## Build, Test, and Development Commands
Use `go build -v ./...` for a full compile check across supported Go 1.24–1.25 toolchains. Run `go test -v ./...` before every push; narrow investigations with `go test ./... -run TestName`. Generate internal scaffolding via `go generate ./...` whenever touching code referenced by `internals_generate.go`. Format and tidy imports with `goimports -local go.mau.fi/whatsmeow -w <file>`; `pre-commit run --all-files` mirrors the CI workflow.

## Coding Style & Naming Conventions
Rely on `gofmt`/`goimports`; tabs are the default indent per `.editorconfig`, with two-space YAML overrides. Match existing zero-values, option structs, and helper names when extending features. Keep exported identifiers descriptive (`Client`, `DangerousInternals`), and prefix experimental or risky APIs with clear warnings. Never edit generated files in `proto/` or `internals.go` by hand—update source definitions and regenerate instead.

## Testing Guidelines
Write table-driven Go tests in files named `*_test.go` and keep them in the same package as the code under test. Assert observable behaviour rather than private state, and cover both happy paths and WhatsApp edge cases (retries, presence updates, media downloads). Aim to keep `go test -cover ./...` stable; add regression tests whenever fixing bugs surfaced in production logs.

## Commit & Pull Request Guidelines
Follow the existing history by using lowercase conventional prefixes (`fix:`, `feat:`, `chore:`) and imperative phrasing (`fix race in prekey upload`). Squash intermediate WIP commits before opening a PR. PRs should explain protocol impacts, reference related issues, and include screenshots or logs when altering client-visible flows. Confirm that `go test ./...` and `pre-commit run --all-files` pass locally before requesting review.

## Security & Configuration Tips
Do not commit real WhatsApp credentials or device states; keep personal `appstate` files out of version control. Sanitise sample payloads added to `store/` or `proto/` and note any redactions in comments. Review dependency bumps for upstream security advisories, and flag breaking protocol changes early so integrators can adapt.
