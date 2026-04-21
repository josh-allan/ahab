# AGENTS.md

## Commands

| Action | Command |
|--------|---------|
| Build | `go build -o ahab ./cmd` |
| Test | `go test ./...` |
| Test single | `go test ./pkg -run TestName` |
| Tidy | `go mod tidy` |

## Architecture

- **Entrypoint:** `cmd/main.go` — cobra CLI with charmbracelet/fang for styling
- **Core logic:** `pkg/ahab.go` — compose operations, file discovery, ignore rules
- **Tests:** `pkg/ahab_test.go` — standard Go testing, table-driven tests

## Key Behavior

- Requires `DOCKER_DIR` env var pointing to the docker compose directory
- Reads `.ahabignore` from `DOCKER_DIR` — supports exact file matches and directory prefixes (trailing `/`)
- `findYAMLFiles` recursively finds `.yaml`/`.yml` files using `filepath.WalkDir`, excluding hidden dirs and `kube/`
- Compose commands run with concurrency limit (`maxConcurrentCommands = 4`) via buffered channel semaphore
- Uses `docker compose` (v2 plugin syntax)
- Errors from parallel goroutines are aggregated via `errors.Join`

## Commands Exposed

`ahab start` · `ahab stop` · `ahab down` · `ahab update` · `ahab restart` · `ahab list`
