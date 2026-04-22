# Ahab

A CLI helper for Docker Compose.

Ahab is a command line tool to manage Docker Compose projects. It discovers compose files recursively under a directory and provides both an interactive TUI and CLI commands for bulk operations.

## Features

- **Interactive TUI**: Run `ahab` with no arguments to launch a bubbletea-based interface for browsing, starting, stopping, and inspecting containers
- **Bulk CLI commands**: Operate on all discovered compose files in parallel (up to 4 at a time)
- **Recursive discovery**: Finds `.yaml` and `.yml` files recursively, skipping hidden directories/files, `kube/`, and `node_modules/`
- **Ignore rules**: Configurable `.ahabignore` file to exclude specific files or directory prefixes

## Usage

### Prerequisites

1. Install dependencies: `go mod tidy`
2. Build the binary: `go build -o ahab ./cmd`
3. Set the `DOCKER_DIR` environment variable to your compose directory
4. Optionally create a `.ahabignore` file in `DOCKER_DIR`

### Interactive TUI (default)

Run `ahab` with no subcommand to launch the TUI:

```bash
ahab
```

Keyboard shortcuts:

| Key | Action |
|-----|--------|
| `j` / `k` or `↑` / `↓` | Navigate files |
| `tab` / `1` / `2` / `3` | Switch pane (info / preview / logs) |
| `s` | Start (`docker compose up -d`) |
| `x` | Stop (`docker compose stop`) |
| `d` | Down (`docker compose down`) |
| `r` | Restart (`docker compose restart`) |
| `p` | Pull (`docker compose pull`) |
| `l` | Toggle logs pane |
| `?` | Toggle help |
| `q` / `ctrl+c` | Quit |

### CLI Commands

```bash
ahab start     # Start all containers (docker compose up -d)
ahab stop      # Stop all containers (docker compose stop)
ahab down      # Stop and remove all resources (docker compose down)
ahab update    # Pull all images (docker compose pull)
ahab restart   # Restart all containers (docker compose restart)
ahab list      # List all discovered compose files (shows ignore status)
```

### Ignore Rules

Create a `.ahabignore` file in `DOCKER_DIR`. Each line is a pattern:

- **Exact file match**: `test1.yaml` — ignores any file named exactly `test1.yaml` anywhere in the tree
- **Directory prefix** (trailing `/`): `home-assistant/` — ignores all files under any directory named `home-assistant`
- Comments start with `#`. Empty lines are ignored.

Example `.ahabignore`:

```
# Ignore specific compose files
backup.yaml
experimental.yml

# Ignore entire directory trees
home-assistant/
```

### Directory Structure

Ahab recursively searches `DOCKER_DIR` for `.yaml` and `.yml` files. It works with any nested structure, for example:

```
docker/
  traefik/
    docker-compose.yaml
  apps/
    home-assistant/
      compose.yaml
    grafana/
      docker-compose.yml
```

### Why?

I wrote this because all of my docker containers are split into discrete modules. This convention helps with organisation and segmentation, however it makes managing containers in bulk tedious.

Solutions such as Ouroborus (no longer maintained), or Watchtower will automatically update containers when a new image is pushed. This carries risk of introducing uncontrolled breaking changes.

This was born to reduce the effort required to manage container updates.
