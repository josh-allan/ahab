package ahab

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/bitfield/script"
)

func readIgnoreFile(dir string) (map[string]struct{}, error) {
	ignorePath := filepath.Join(dir, ".ahabignore")
	file, err := os.Open(ignorePath)
	if err != nil {
		// If the file doesn't exist, just return an empty set
		if os.IsNotExist(err) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	defer file.Close()

	ignores := make(map[string]struct{})
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ignores[line] = struct{}{}
	}
	return ignores, scanner.Err()
}

func getDockerFiles(dir string) *script.Pipe {
	return script.FindFiles(dir).
		MatchRegexp(regexp.MustCompile(`\.ya?ml$`)).
		RejectRegexp(regexp.MustCompile(`/(?:\.[^/]+)/`)). // this is because hidden directories could have yaml in them which doesn't compile and then breaks
		RejectRegexp(regexp.MustCompile(`/kube(/|$)`))     // similarly, we don't want to run docker-compose on kube files
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

func runComposeCommands(args ...string) *script.Pipe {
	return script.Exec(strings.Join(args, " "))
}

func findComposeFiles(action string) ([]string, error) {
	dir, err := getDockerDir()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Finding Dockerfiles to %s...\n", action)
	dockerFiles := getDockerFiles(dir)
	output, err := dockerFiles.String()
	if err != nil {
		return nil, err
	}

	files := strings.Fields(output)
	if len(files) == 0 {
		fmt.Printf("No docker-compose files found to %s.\n", action)
		return nil, nil
	}

	ignores, err := readIgnoreFile(dir)
	if err != nil {
		return nil, err
	}

	var filtered []string
	for _, file := range files {
		if _, skip := ignores[file]; skip {
			fmt.Printf("Skipping %s (ignored by .ahabignore)\n", file)
			continue
		}
		filtered = append(filtered, file)
	}

	return filtered, nil
}

func ListIgnoreFiles() error {
	files, err := findComposeFiles("--dry-run start")
	if err != nil || len(files) == 0 {
		return err
	}
	fmt.Println("Listing docker-compose files...")
	for _, file := range files {
		fmt.Println(file)
	}
	return nil
}

func runOnFiles(files []string, action string, cmdArgs []string) {
	fmt.Printf("%s docker-compose for each file...\n", action)
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			args := append([]string{"docker-compose", "-f", f}, cmdArgs...)
			fmt.Printf("Running: %s\n", strings.Join(args, " "))
			_, err := runComposeCommands(args...).Stdout()
			if err != nil {
				fmt.Printf("Error running %s for %s: %v\n", action, f, err)
			}
		}(file)
	}

	wg.Wait()
}

func runAction(action string, cmdArgs ...string) error {
	files, err := findComposeFiles(action)
	if err != nil || len(files) == 0 {
		return err
	}
	runOnFiles(files, action, cmdArgs)
	return nil
}

func RunAllCompose() error     { return runAction("start", "up", "-d") }
func UpdateAllCompose() error  { return runAction("update", "pull") }
func StopAllCompose() error    { return runAction("stop", "stop") }
func RestartAllCompose() error { return runAction("restart", "restart") }
