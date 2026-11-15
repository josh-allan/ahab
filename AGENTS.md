# AGENTS.md

## Commands

| Action | Command |
|--------|---------|
| Build | `go build -o ahab ./cmd` (add `-buildvcs=false` if git rejects the checkout's ownership) |
| Test | `go test ./...` |
| Test single | `go test ./pkg -run TestName` |
| Tidy | `go mod tidy` |

## Architecture

- **Entrypoint:** `cmd/main.go` -- cobra + charmbracelet/fang; prints help with no arguments; `ahab tui` launches the TUI
- **Catalogue:** `pkg/catalogue.go` -- `Stack` (context + compose file), `.ahabstacks` parsing, target resolution
- **Snapshot:** `pkg/snapshot.go` -- builds `.ahabstacks` from the compose label `com.docker.compose.project.config_files` on running containers in every `docker context ls` entry
- **Actions:** `pkg/ahab.go` -- `Action` (name + ordered compose steps), `Action.RunOn` (one stack), `Run` (bulk), `List`, `Status`
- **TUI:** `internal/tui/` -- bubbletea; lists catalogued stacks, keys map to `ahab.Action`
- **Tests:** table-driven, `t.TempDir` fixtures; no docker needed

## Key Behavior

- `DOCKER_DIR` must point at the compose root; `.ahabstacks` lives there
- Bulk commands act only on `.ahabstacks`; a missing catalogue or a listed file that does not exist is an error (fail closed)
- Explicit files (`ahab start <file>`) take their context from `--context`, then the catalogue, then docker's current context
- `update` = `pull --ignore-buildable`, `build --pull`, `up -d`; the `up -d` step is what makes a reboot come back on the new image
- At most 4 stacks run concurrently; per-stack errors are joined with `errors.Join`

## Commands Exposed

`ahab start` · `stop` · `down` · `restart` · `update` · `list` · `snapshot` · `tui`

## Style Guide

All Go code follows [Effective Go](https://go.dev/doc/effective_go) and the [Google Go Style Guide](https://google.github.io/styleguide/go/). Run `go vet ./...` and `gofmt -l .` before committing.
