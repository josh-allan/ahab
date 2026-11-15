# Ahab

Bulk Docker Compose operations over an explicit list of stacks.

## Setup

```fish
go build -o ~/.local/bin/ahab ./cmd
set -Ux DOCKER_DIR /path/to/compose/root
```

## `.ahabstacks`

Bulk commands act only on the stacks listed in `$DOCKER_DIR/.ahabstacks`, one
`<docker context> <compose file>` per line (paths relative to `DOCKER_DIR`, `#` comments).
`ahab snapshot` writes it from the running stacks. A listed file that does not exist is an error.

```
rootless byteandbuild.org/caddy/compose.yaml
default  byteandbuild.org/homeassistant/compose.yaml
```

## Commands

| Command | Does |
|---|---|
| `ahab update` | `pull --ignore-buildable`, `build --pull`, `up -d` |
| `ahab start` / `stop` / `restart` / `down` | the matching compose command |
| `ahab snapshot` | regenerate `.ahabstacks` from what is running in every context |
| `ahab list` | print the catalogue |
| `ahab tui` | interactive browser (`?` for keys) |

Pass compose files to act on them instead of the catalogue: `ahab start x/compose.yaml`.
The context comes from `--context`, then the catalogue entry, then docker's current context.
