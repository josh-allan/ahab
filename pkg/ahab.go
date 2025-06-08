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

func RunAllCompose() error {
	dir, err := GetDockerDir()
	if err != nil {
		return err
	}

	fmt.Println("Finding Dockerfiles...")
	dockerFiles := GetDockerFiles(dir)
	output, err := dockerFiles.String()
	if err != nil {
		return err
	}

	files := strings.Fields(output)
	if len(files) == 0 {
		fmt.Println("No docker-compose files found.")
		return nil
	}

	StartComposeFiles(files)
	return nil
}

func UpdateAllCompose() error {
	dir, err := GetDockerDir()
	if err != nil {
		return err
	}

	fmt.Println("Finding Dockerfiles for update...")
	dockerFiles := GetDockerFiles(dir)
	output, err := dockerFiles.String()
	if err != nil {
		return err
	}

	files := strings.Fields(output)
	if len(files) == 0 {
		fmt.Println("No docker-compose files found for update.")
		return nil
	}

	UpdateComposeFiles(files)
	return nil
}
