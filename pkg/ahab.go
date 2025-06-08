package ahab

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bitfield/script"
)


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

    return files, nil
}

func startComposeFiles(files []string) {
	fmt.Println("Starting docker-compose for each file...")
	for _, file := range files {
		fmt.Printf("Running: docker-compose -f %s up -d\n", file)
		_, err := runComposeCommands("docker-compose", "-f", file, "up", "-d").Stdout()
		if err != nil {
			fmt.Printf("Error running docker-compose for %s: %v\n", file, err)
		}
	}
}

func updateComposeFiles(files []string) {
	fmt.Println("Updating docker-compose for each file...")
	for _, file := range files {
		fmt.Printf("Running: docker-compose -f %s pull\n", file)
		_, err := runComposeCommands("docker-compose", "-f", file, "pull").Stdout()
		if err != nil {
			fmt.Printf("Error updating docker-compose for %s: %v\n", file, err)
		}
	}
}

func stopComposeFiles(files []string) {
	fmt.Println("Stopping docker-compose for each file...")
	for _, file := range files {
		fmt.Printf("Running: docker-compose -f %s down\n", file)
		_, err := runComposeCommands("docker-compose", "-f", file, "down").Stdout()
		if err != nil {
			fmt.Printf("Error stopping docker-compose for %s: %v\n", file, err)
		}
	}
}

func restartComposeFiles(files []string) {
	fmt.Println("Restarting docker-compose for each file...")
	for _, file := range files {
		fmt.Printf("Running: docker-compose -f %s restart\n", file)
		_, err := runComposeCommands("docker-compose", "-f", file, "restart").Stdout()
		if err != nil {
			fmt.Printf("Error restarting docker-compose for %s: %v\n", file, err)
		}
	}
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