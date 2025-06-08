package ahab

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bitfield/script"
)


func GetDockerFiles(dir string) *script.Pipe {
	return script.FindFiles(dir).
		MatchRegexp(regexp.MustCompile(`\.ya?ml$`)).
		RejectRegexp(regexp.MustCompile(`/(?:\.[^/]+)/`)). // this is because hidden directories could have yaml in them which doesn't compile and then breaks
		RejectRegexp(regexp.MustCompile(`/kube(/|$)`))     // similarly, we don't want to run docker-compose on kube files
}

func GetDockerDir() (string, error) {
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

func FindComposeFiles(action string) ([]string, error) {
    dir, err := GetDockerDir()
    if err != nil {
        return nil, err
    }

    fmt.Printf("Finding Dockerfiles to %s...\n", action)
    dockerFiles := GetDockerFiles(dir)
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

func StartComposeFiles(files []string) {
	fmt.Println("Starting docker-compose for each file...")
	for _, file := range files {
		fmt.Printf("Running: docker-compose -f %s up -d\n", file)
		_, err := runComposeCommands("docker-compose", "-f", file, "up", "-d").Stdout()
		if err != nil {
			fmt.Printf("Error running docker-compose for %s: %v\n", file, err)
		}
	}
}

func UpdateComposeFiles(files []string) {
	fmt.Println("Updating docker-compose for each file...")
	for _, file := range files {
		fmt.Printf("Running: docker-compose -f %s pull\n", file)
		_, err := runComposeCommands("docker-compose", "-f", file, "pull").Stdout()
		if err != nil {
			fmt.Printf("Error updating docker-compose for %s: %v\n", file, err)
		}
	}
}

func StopComposeFiles(files []string) {
	fmt.Println("Stopping docker-compose for each file...")
	for _, file := range files {
		fmt.Printf("Running: docker-compose -f %s down\n", file)
		_, err := runComposeCommands("docker-compose", "-f", file, "down").Stdout()
		if err != nil {
			fmt.Printf("Error stopping docker-compose for %s: %v\n", file, err)
		}
	}
}

func RestartComposeFiles(files []string) {
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
	files, err := FindComposeFiles("start")
	if err != nil || len(files) == 0 {
		return err
	}
	StartComposeFiles(files)
	return nil
}

func UpdateAllCompose() error {
	files, err := FindComposeFiles("update")
	if err != nil || len(files) == 0 {
		return err
	}

	UpdateComposeFiles(files)
	return nil
}

func StopAllCompose() error {
	files, err := FindComposeFiles("stop")
	if err != nil || len(files) == 0 {
		return err
	}

	StopComposeFiles(files)
	return nil
}

func RestartAllCompose() error {
	files, err := FindComposeFiles("restart")
	if err != nil || len(files) == 0 {
		return err
	}

	RestartComposeFiles(files)
	return nil
}