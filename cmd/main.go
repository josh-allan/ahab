package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitfield/script"
)




func main() {

	dir, err := ahab.getDockerDir()
	if err != nil {
	fmt.Println("Error:", err)
	os.Exit(1)
	}

	fmt.Println("Finding Dockerfiles...")
	script.FindFiles(dir)

	fmt.Println("We only care about the files that end with .yaml or .yml")

    dockerFiles := getDockerFiles(dir)
    output, err := dockerFiles.String()
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    files := strings.Fields(output) // splits on whitespace/newlines

}
