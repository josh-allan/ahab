package ahab

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func Test_ignoreRules_match(t *testing.T) {
	rules := ignoreRules{
		exact: map[string]struct{}{
			"/docker/test1.yaml": {},
		},
		prefixes: []string{
			"/docker/home-assistant/",
		},
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "exact match",
			path: "/docker/test1.yaml",
			want: true,
		},
		{
			name: "prefix match",
			path: "/docker/home-assistant/docker.yaml",
			want: true,
		},
		{
			name: "no match",
			path: "/docker/valid-service.yaml",
			want: false,
		},
		{
			name: "empty rules",
			path: "/docker/anything.yaml",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rules.match(tt.path); got != tt.want {
				t.Errorf("ignoreRules.match(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func Test_readIgnoreFile(t *testing.T) {
	tests := []struct {
		name    string
		dir     string
		want    ignoreRules
		wantErr bool
	}{
		{
			name: "valid directory with patterns",
			dir:  "./testdata",
			want: ignoreRules{
				exact: map[string]struct{}{
					"test1.yaml": {},
					"test2.yaml": {},
				},
				prefixes: []string{
					"home-assistant/",
				},
			},
			wantErr: false,
		},
		{
			name:    "non-existent directory returns empty rules",
			dir:     "./non_existent_dir",
			want:    ignoreRules{exact: map[string]struct{}{}, prefixes: nil},
			wantErr: false,
		},
		{
			name:    "empty directory returns empty rules",
			dir:     "./testdata/empty",
			want:    ignoreRules{exact: map[string]struct{}{}, prefixes: nil},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readIgnoreFile(tt.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("readIgnoreFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.exact, tt.want.exact) {
				t.Errorf("readIgnoreFile() exact = %v, want %v", got.exact, tt.want.exact)
			}
			if !reflect.DeepEqual(got.prefixes, tt.want.prefixes) {
				t.Errorf("readIgnoreFile() prefixes = %v, want %v", got.prefixes, tt.want.prefixes)
			}
		})
	}
}

func Test_findYAMLFiles(t *testing.T) {
	tests := []struct {
		name      string
		dir       string
		wantNames []string
		wantErr   bool
	}{
		{
			name:      "finds yaml and yml files recursively, excludes hidden dirs/files, kube, and node_modules but includes dirs with dots in their names",
			dir:       "./testdata",
			wantNames: []string{"another-service.yml", "compose.yaml", "docker.yaml", "valid-service.yaml"},
			wantErr:   false,
		},
		{
			name:    "non-existent directory returns error",
			dir:     "./non_existent_dir",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findYAMLFiles(tt.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("findYAMLFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			var gotNames []string
			for _, f := range got {
				gotNames = append(gotNames, filepath.Base(f))
			}
			sort.Strings(gotNames)
			wantSorted := make([]string, len(tt.wantNames))
			copy(wantSorted, tt.wantNames)
			sort.Strings(wantSorted)
			if !reflect.DeepEqual(gotNames, wantSorted) {
				t.Errorf("findYAMLFiles() names = %v, want %v", gotNames, wantSorted)
			}
		})
	}
}

func Test_getDockerDir(t *testing.T) {
	tests := []struct {
		name    string
		envVal  string
		want    string
		wantErr bool
	}{
		{
			name:    "unset DOCKER_DIR returns error",
			envVal:  "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid path returns error",
			envVal:  "/nonexistent/path/that/does/not/exist",
			want:    "",
			wantErr: true,
		},
		{
			name:    "valid directory returns path",
			envVal:  "./testdata",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := os.Getenv("DOCKER_DIR")
			defer os.Setenv("DOCKER_DIR", orig)
			os.Setenv("DOCKER_DIR", tt.envVal)
			got, err := getDockerDir()
			if (err != nil) != tt.wantErr {
				t.Errorf("getDockerDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got == "" {
				t.Errorf("getDockerDir() returned empty string for valid dir")
			}
		})
	}
}
