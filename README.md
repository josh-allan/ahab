# Ahab
A CLI helper for Docker Compose. 

Ahab is a simple command line tool to manage Docker Compose projects. It has four primary functions:

- Start all containers
- Stop all containers
- Update all containers
- Restart all containers.
- List all containers (note this is a dry run command, purely for testing).

### Usage

Usage is simple:

1. `go mod tidy`
2. Compile the binary: `go build -o ahab ./cmd`
3. Set `DOCKER_DIR` environment variable.
4. Enjoy!

### Why?

I wrote this because all of my docker containers are split into discrete modules. They also adhere to the following directory and naming convention:

```
- /docker/<container_name_directory>/<container_name.yaml>
```

This convention massively helps with organisation and segmentation, however it makes managing containers in bulk tedious.

Solutions such as Ouroborus (no longer maintained), or Watchtower will automatically update containers when a new image is pushed. This carries risk of introducing uncontrolled breaking changes. 

This was born to reduce the effort required to manage container updates.

Note: The ignore file expects absolute paths.
