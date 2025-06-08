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
		RejectRegexp(regexp.MustCompile(`/kube(/|$)`)) // similarly, we don't want to run docker-compose on kube files
}

func GetDockerDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get home directory: %w", err)
	}
	dir := home + "/dev/docker"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return "", fmt.Errorf("docker directory does not exist: %s", dir)
	}
	return dir, nil 
}

func RunComposeCommands(args ...string) *script.Pipe {
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