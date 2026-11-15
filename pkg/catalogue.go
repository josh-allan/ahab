package ahab

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const catalogueFile = ".ahabstacks"

// ErrNoCatalogue means DOCKER_DIR has no .ahabstacks file.
var ErrNoCatalogue = errors.New("no " + catalogueFile + " catalogue")

// Stack is one compose file and the docker context it runs under.
// An empty Context means docker's current context.
type Stack struct {
	Context string
	File    string
}

// readCatalogue parses dir/.ahabstacks: one "<context> <compose file>" per
// line, relative paths resolved against dir. A listed file that does not
// exist is an error, so path drift fails loudly.
func readCatalogue(dir string) ([]Stack, error) {
	path := filepath.Join(dir, catalogueFile)
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%s: %w", path, ErrNoCatalogue)
	}
	if err != nil {
		return nil, err
	}

	var stacks []Stack
	for i, line := range strings.Split(string(data), "\n") {
		n := i + 1
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("%s:%d: want \"<context> <compose file>\", got %q", path, n, line)
		}
		file := fields[1]
		if !filepath.IsAbs(file) {
			file = filepath.Join(dir, file)
		}
		if _, err := os.Stat(file); err != nil {
			return nil, fmt.Errorf("%s:%d: %w", path, n, err)
		}
		stacks = append(stacks, Stack{Context: fields[0], File: file})
	}
	return stacks, nil
}

// LoadCatalogue reads the catalogue in DOCKER_DIR.
func LoadCatalogue() ([]Stack, error) {
	dir, err := getDockerDir()
	if err != nil {
		return nil, err
	}
	return readCatalogue(dir)
}

// resolveTargets returns the whole catalogue when files is empty. Otherwise
// it returns one Stack per file, with context from contextFlag, else the
// file's catalogue entry, else docker's current context.
func resolveTargets(dir string, files []string, contextFlag string) ([]Stack, error) {
	catalogue, err := readCatalogue(dir)
	if len(files) == 0 {
		return catalogue, err
	}
	if err != nil && !errors.Is(err, ErrNoCatalogue) {
		return nil, err
	}

	listed := make(map[string]string, len(catalogue))
	for _, s := range catalogue {
		listed[s.File] = s.Context
	}

	stacks := make([]Stack, 0, len(files))
	for _, f := range files {
		abs, err := filepath.Abs(f)
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(abs); err != nil {
			return nil, err
		}
		ctx := contextFlag
		if ctx == "" {
			ctx = listed[abs]
		}
		stacks = append(stacks, Stack{Context: ctx, File: abs})
	}
	return stacks, nil
}
