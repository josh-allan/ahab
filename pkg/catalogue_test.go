package ahab

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReadCatalogue(t *testing.T) {
	dir := t.TempDir()
	rel := filepath.Join(dir, "a", "compose.yaml")
	abs := filepath.Join(dir, "b", "b.yml")
	writeFile(t, rel, "services: {}\n")
	writeFile(t, abs, "services: {}\n")
	writeFile(t, filepath.Join(dir, catalogueFile),
		"# comment\n\nrootless a/compose.yaml\ndefault   "+abs+"\n")

	got, err := readCatalogue(dir)
	if err != nil {
		t.Fatalf("readCatalogue: %v", err)
	}
	want := []Stack{{Context: "rootless", File: rel}, {Context: "default", File: abs}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("readCatalogue = %+v, want %+v", got, want)
	}
}

func TestReadCatalogueErrors(t *testing.T) {
	tests := []struct {
		name      string
		content   string // "" = no catalogue file
		wantNoCat bool
	}{
		{name: "missing catalogue", wantNoCat: true},
		{name: "one field", content: "a/compose.yaml\n"},
		{name: "three fields", content: "rootless a/compose.yaml extra\n"},
		{name: "file does not exist", content: "rootless gone/compose.yaml\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, filepath.Join(dir, "a", "compose.yaml"), "services: {}\n")
			if tt.content != "" {
				writeFile(t, filepath.Join(dir, catalogueFile), tt.content)
			}
			_, err := readCatalogue(dir)
			if err == nil {
				t.Fatal("readCatalogue: want error, got nil")
			}
			if got := errors.Is(err, ErrNoCatalogue); got != tt.wantNoCat {
				t.Errorf("errors.Is(err, ErrNoCatalogue) = %v, want %v (err: %v)", got, tt.wantNoCat, err)
			}
		})
	}
}

func TestResolveTargets(t *testing.T) {
	dir := t.TempDir()
	listed := filepath.Join(dir, "a", "compose.yaml")
	unlisted := filepath.Join(dir, "c", "compose.yaml")
	writeFile(t, listed, "services: {}\n")
	writeFile(t, unlisted, "services: {}\n")
	writeFile(t, filepath.Join(dir, catalogueFile), "default a/compose.yaml\n")

	tests := []struct {
		name        string
		files       []string
		contextFlag string
		want        []Stack
	}{
		{name: "no files: whole catalogue", want: []Stack{{Context: "default", File: listed}}},
		{name: "listed file keeps catalogue context", files: []string{listed}, want: []Stack{{Context: "default", File: listed}}},
		{name: "flag overrides catalogue", files: []string{listed}, contextFlag: "rootless", want: []Stack{{Context: "rootless", File: listed}}},
		{name: "unlisted file uses current context", files: []string{unlisted}, want: []Stack{{File: unlisted}}},
		{name: "unlisted file with flag", files: []string{unlisted}, contextFlag: "default", want: []Stack{{Context: "default", File: unlisted}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveTargets(dir, tt.files, tt.contextFlag)
			if err != nil {
				t.Fatalf("resolveTargets: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("resolveTargets = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestResolveTargetsWithoutCatalogue(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "c", "compose.yaml")
	writeFile(t, file, "services: {}\n")

	if _, err := resolveTargets(dir, nil, ""); !errors.Is(err, ErrNoCatalogue) {
		t.Errorf("bulk without catalogue: err = %v, want ErrNoCatalogue", err)
	}
	got, err := resolveTargets(dir, []string{file}, "")
	if err != nil {
		t.Fatalf("explicit file without catalogue: %v", err)
	}
	if want := []Stack{{File: file}}; !reflect.DeepEqual(got, want) {
		t.Errorf("resolveTargets = %+v, want %+v", got, want)
	}
}
