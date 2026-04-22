package ahab

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var yamlRegex = regexp.MustCompile(`\.ya?ml$`)

const maxConcurrentCommands = 4

type ignoreRules struct {
	exact    map[string]struct{}
	prefixes []string
}

func (r ignoreRules) match(path string) bool {
	if _, ok := r.exact[path]; ok {
		return true
	}
	for _, prefix := range r.prefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

func readIgnoreFile(dir string) (ignoreRules, error) {
	ignorePath := filepath.Join(dir, ".ahabignore")
	file, err := os.Open(ignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ignoreRules{exact: map[string]struct{}{}}, nil
		}
		return ignoreRules{}, err
	}
	defer file.Close()

	rules := ignoreRules{exact: make(map[string]struct{})}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasSuffix(line, "/") {
			rules.prefixes = append(rules.prefixes, line)
		} else {
			rules.exact[line] = struct{}{}
		}
	}
	return rules, scanner.Err()
}

func getDockerDir() (string, error) {
	if dir := os.Getenv("DOCKER_DIR"); dir != "" {
		if _, err := os.Stat(dir); err == nil {
			return dir, nil
		}
		return "", fmt.Errorf("docker directory from DOCKER_DIR does not exist: %s", dir)
	}
	return "", fmt.Errorf("DOCKER_DIR environment variable is not set")
}

func findYAMLFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") || d.Name() == "kube" || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}
		if yamlRegex.MatchString(d.Name()) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func findComposeFiles(action string) ([]string, error) {
	dir, err := getDockerDir()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Finding Docker Compose files to %s...\n", action)
	files, err := findYAMLFiles(dir)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		fmt.Printf("No Docker Compose files found to %s.\n", action)
		return []string{}, nil
	}

	rules, err := readIgnoreFile(dir)
	if err != nil {
		return nil, err
	}

	var filtered []string
	for _, file := range files {
		if rules.match(file) {
			fmt.Printf("Skipping %s (ignored by .ahabignore)\n", file)
			continue
		}
		filtered = append(filtered, file)
	}

	return filtered, nil
}

// ComposeFileInfo holds info about a compose file for the TUI.
type ComposeFileInfo struct {
	Path    string
	Ignored bool
}

// FindComposeFilesForTUI returns all YAML files in DOCKER_DIR, with ignore status.
func FindComposeFilesForTUI() ([]ComposeFileInfo, error) {
	dir, err := getDockerDir()
	if err != nil {
		return nil, err
	}
	files, err := findYAMLFiles(dir)
	if err != nil {
		return nil, err
	}
	rules, err := readIgnoreFile(dir)
	if err != nil {
		return nil, err
	}
	var result []ComposeFileInfo
	for _, f := range files {
		result = append(result, ComposeFileInfo{
			Path:    f,
			Ignored: rules.match(f),
		})
	}
	return result, nil
}

// GetComposeStatus runs docker compose ps and returns "running", "stopped", "partial", or "unknown".
func GetComposeStatus(file string) string {
	cmd := exec.Command("docker", "compose", "-f", file, "ps", "--format", "json")
	out, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || lines[0] == "" || lines[0] == "[]" {
		return "stopped"
	}

	type containerPs struct {
		State string `json:"State"`
	}

	var running, total int
	for _, line := range lines {
		if line == "" || line == "[]" {
			continue
		}
		var c containerPs
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			continue
		}
		total++
		if c.State == "running" {
			running++
		}
	}
	if total == 0 {
		return "stopped"
	}
	if running == total {
		return "running"
	}
	if running == 0 {
		return "stopped"
	}
	return "partial"
}

// ExecCompose runs docker compose on a single file with given args.
func ExecCompose(ctx context.Context, stdout, stderr io.Writer, file string, args ...string) error {
	return execCompose(ctx, stdout, stderr, file, args...)
}

func execCompose(ctx context.Context, stdout, stderr io.Writer, file string, args ...string) error {
	cmdArgs := append([]string{"compose", "-f", file}, args...)
	cmd := exec.CommandContext(ctx, "docker", cmdArgs...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	fmt.Printf("Running: docker %s\n", strings.Join(cmdArgs, " "))
	return cmd.Run()
}

func runOnFiles(ctx context.Context, files []string, action string, cmdArgs []string) error {
	fmt.Printf("%s docker compose for each file...\n", action)
	sem := make(chan struct{}, maxConcurrentCommands)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	for _, file := range files {
		sem <- struct{}{}
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			defer func() { <-sem }()
			if err := execCompose(ctx, os.Stdout, os.Stderr, f, cmdArgs...); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("%s: %w", f, err))
				mu.Unlock()
			}
		}(file)
	}

	wg.Wait()
	return errors.Join(errs...)
}

func runAction(action string, cmdArgs ...string) error {
	ctx := context.Background()
	files, err := findComposeFiles(action)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}
	return runOnFiles(ctx, files, action, cmdArgs)
}

func RunAllCompose() error      { return runAction("start", "up", "-d") }
func UpdateAllCompose() error   { return runAction("update", "pull") }
func StopAllCompose() error     { return runAction("stop", "stop") }
func StopAllComposeDown() error { return runAction("down", "down") }
func RestartAllCompose() error  { return runAction("restart", "restart") }

func ListIgnoreFiles() error {
	dir, err := getDockerDir()
	if err != nil {
		return err
	}

	fmt.Println("Listing Docker Compose files...")
	files, err := findYAMLFiles(dir)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		fmt.Println("No Docker Compose files found.")
		return nil
	}

	rules, err := readIgnoreFile(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if rules.match(file) {
			fmt.Printf("  [ignored] %s\n", file)
		} else {
			fmt.Printf("  %s\n", file)
		}
	}
	return nil
}
