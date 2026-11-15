package ahab

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestUniqueLines(t *testing.T) {
	out := []byte("/d/b/compose.yaml\n\n/d/a/compose.yaml\n/d/b/compose.yaml\n")
	want := []string{"/d/a/compose.yaml", "/d/b/compose.yaml"}
	if got := uniqueLines(out); !reflect.DeepEqual(got, want) {
		t.Errorf("uniqueLines = %q, want %q", got, want)
	}
}

func TestFormatCatalogueRoundTrip(t *testing.T) {
	dir := t.TempDir()
	inside := filepath.Join(dir, "a", "compose.yaml")
	outside := filepath.Join(t.TempDir(), "x.yml")
	writeFile(t, inside, "services: {}\n")
	writeFile(t, outside, "services: {}\n")
	stacks := []Stack{{Context: "rootless", File: inside}, {Context: "default", File: outside}}

	text := formatCatalogue(dir, stacks)
	if err := os.WriteFile(filepath.Join(dir, catalogueFile), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := readCatalogue(dir)
	if err != nil {
		t.Fatalf("readCatalogue: %v\n%s", err, text)
	}
	if !reflect.DeepEqual(got, stacks) {
		t.Errorf("round trip = %+v, want %+v\n%s", got, stacks, text)
	}
}
