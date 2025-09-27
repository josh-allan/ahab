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
		rel, _ := filepath.Rel(dir, file)
		if _, skip := ignores[rel]; skip {
			fmt.Printf("Skipping %s (ignored by .ahabignore)\n", rel)
			continue
		}
		filtered = append(filtered, file)
	}

	return filtered, nil
}

func startComposeFiles(files []string) {
	fmt.Println("Starting docker-compose for each file...")
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			fmt.Printf("Running: docker-compose -f %s up -d\n", f)
			_, err := runComposeCommands("docker-compose", "-f", f, "up", "-d").Stdout()
			if err != nil {
				fmt.Printf("Error running docker-compose for %s: %v\n", f, err)
			}
		}(file)
	}

	wg.Wait()
}

func updateComposeFiles(files []string) {
	fmt.Println("Updating docker-compose for each file...")
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			fmt.Printf("Running: docker-compose -f %s pull\n", f)
			_, err := runComposeCommands("docker-compose", "-f", f, "pull").Stdout()
			if err != nil {
				fmt.Printf("Error updating docker-compose for %s: %v\n", f, err)
			}
		}(file)
	}

	wg.Wait()
}

func stopComposeFiles(files []string) {
	fmt.Println("Stopping docker-compose for each file...")
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			fmt.Printf("Running: docker-compose -f %s stop\n", f)
			_, err := runComposeCommands("docker-compose", "-f", f, "stop").Stdout()
			if err != nil {
				fmt.Printf("Error stopping docker-compose for %s: %v\n", f, err)
			}
		}(file)
	}

	wg.Wait()
}

func restartComposeFiles(files []string) {
	fmt.Println("Restarting docker-compose for each file...")
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			fmt.Printf("Running: docker-compose -f %s restart\n", f)
			_, err := runComposeCommands("docker-compose", "-f", f, "restart").Stdout()
			if err != nil {
				fmt.Printf("Error restarting docker-compose for %s: %v\n", f, err)
			}
		}(file)
	}

	wg.Wait()
}

func RunAllCompose() error {
	files, err := findComposeFiles("start")
	if err != nil || len(files) == 0 {
		return err
	}
	startComposeFiles(files)
	return nil
}

func UpdateAllCompose() error {
	files, err := findComposeFiles("update")
	if err != nil || len(files) == 0 {
		return err
	}

	updateComposeFiles(files)
	return nil
}

func StopAllCompose() error {
	files, err := findComposeFiles("stop")
	if err != nil || len(files) == 0 {
		return err
	}

	stopComposeFiles(files)
	return nil
}

func RestartAllCompose() error {
	files, err := findComposeFiles("restart")
	if err != nil || len(files) == 0 {
		return err
	}

	restartComposeFiles(files)
	return nil
}
