// Package ahab runs docker compose operations across a catalogue of stacks.
package ahab

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
)

const maxConcurrentCommands = 4

// Action is a named operation: compose argument lists run in order per stack.
type Action struct {
	Name  string
	Steps [][]string
}

var (
	Start = Action{Name: "start", Steps: [][]string{{"up", "-d"}}}
	// Update pulls registry images, rebuilds local ones against fresh base
	// images, then recreates changed containers. A container keeps the image
	// it was created from, so only the `up -d` step makes a reboot come back
	// on the new version.
	Update  = Action{Name: "update", Steps: [][]string{{"pull", "--ignore-buildable"}, {"build", "--pull"}, {"up", "-d"}}}
	Stop    = Action{Name: "stop", Steps: [][]string{{"stop"}}}
	Down    = Action{Name: "down", Steps: [][]string{{"down"}}}
	Restart = Action{Name: "restart", Steps: [][]string{{"restart"}}}
)

func getDockerDir() (string, error) {
	dir := os.Getenv("DOCKER_DIR")
	if dir == "" {
		return "", errors.New("DOCKER_DIR environment variable is not set")
	}
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("DOCKER_DIR: %w", err)
	}
	return dir, nil
}

// ComposeArgs returns the docker argv that runs compose args against s.
func ComposeArgs(s Stack, args ...string) []string {
	var out []string
	if s.Context != "" {
		out = append(out, "--context", s.Context)
	}
	out = append(out, "compose", "-f", s.File)
	return append(out, args...)
}

// RunOn runs a's steps on s in order, stopping at the first failure.
func (a Action) RunOn(ctx context.Context, stdout, stderr io.Writer, s Stack) error {
	for _, step := range a.Steps {
		argv := ComposeArgs(s, step...)
		if _, err := fmt.Fprintf(stdout, "Running: docker %s\n", strings.Join(argv, " ")); err != nil {
			return err
		}
		cmd := exec.CommandContext(ctx, "docker", argv...)
		cmd.Stdout, cmd.Stderr = stdout, stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %s: %w", s.File, strings.Join(step, " "), err)
		}
	}
	return nil
}

// Run applies a to the catalogue, or to files when given, at most
// maxConcurrentCommands stacks at a time. contextFlag overrides each
// target's docker context.
func Run(a Action, files []string, contextFlag string) error {
	dir, err := getDockerDir()
	if err != nil {
		return err
	}
	stacks, err := resolveTargets(dir, files, contextFlag)
	if err != nil {
		return err
	}

	fmt.Printf("Running %s on %d stack(s)...\n", a.Name, len(stacks))
	sem := make(chan struct{}, maxConcurrentCommands)
	errs := make([]error, len(stacks))
	var wg sync.WaitGroup
	for i, s := range stacks {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			errs[i] = a.RunOn(context.Background(), os.Stdout, os.Stderr, s)
		}()
	}
	wg.Wait()
	return errors.Join(errs...)
}

// List prints each catalogued stack with its context.
func List() error {
	stacks, err := LoadCatalogue()
	if err != nil {
		return err
	}
	for _, s := range stacks {
		fmt.Printf("%-10s %s\n", s.Context, s.File)
	}
	return nil
}

// Status reports "running", "partial", "stopped", or "unknown" for s.
func Status(s Stack) string {
	out, err := exec.Command("docker", ComposeArgs(s, "ps", "--all", "--format", "json")...).Output()
	if err != nil {
		return "unknown"
	}
	return statusOf(out)
}

// statusOf classifies `compose ps --format json` output (one object per line).
func statusOf(out []byte) string {
	var running, total int
	for _, line := range strings.Split(string(out), "\n") {
		var c struct{ State string }
		if json.Unmarshal([]byte(line), &c) != nil {
			continue
		}
		total++
		if c.State == "running" {
			running++
		}
	}
	switch running {
	case 0:
		return "stopped"
	case total:
		return "running"
	default:
		return "partial"
	}
}
