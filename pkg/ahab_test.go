package ahab

import (
	"reflect"
	"testing"
)

func TestGetDockerDir(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		wantErr bool
	}{
		{name: "unset", env: "", wantErr: true},
		{name: "missing dir", env: "/nonexistent/ahab/dir", wantErr: true},
		{name: "valid dir", env: t.TempDir()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DOCKER_DIR", tt.env)
			got, err := getDockerDir()
			if (err != nil) != tt.wantErr {
				t.Fatalf("getDockerDir() err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.env {
				t.Errorf("getDockerDir() = %q, want %q", got, tt.env)
			}
		})
	}
}

func TestComposeArgs(t *testing.T) {
	tests := []struct {
		name  string
		stack Stack
		args  []string
		want  []string
	}{
		{
			name:  "with context",
			stack: Stack{Context: "default", File: "/d/ha/compose.yaml"},
			args:  []string{"up", "-d"},
			want:  []string{"--context", "default", "compose", "-f", "/d/ha/compose.yaml", "up", "-d"},
		},
		{
			name:  "current context",
			stack: Stack{File: "/d/x.yaml"},
			args:  []string{"pull"},
			want:  []string{"compose", "-f", "/d/x.yaml", "pull"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ComposeArgs(tt.stack, tt.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ComposeArgs = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUpdateRecreatesAfterPull(t *testing.T) {
	want := [][]string{{"pull", "--ignore-buildable"}, {"build", "--pull"}, {"up", "-d"}}
	if !reflect.DeepEqual(Update.Steps, want) {
		t.Errorf("Update.Steps = %q, want %q", Update.Steps, want)
	}
}

func TestStatusOf(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want string
	}{
		{name: "no containers", out: "", want: "stopped"},
		{name: "empty array", out: "[]\n", want: "stopped"},
		{name: "all running", out: `{"State":"running"}` + "\n" + `{"State":"running"}` + "\n", want: "running"},
		{name: "some running", out: `{"State":"running"}` + "\n" + `{"State":"exited"}` + "\n", want: "partial"},
		{name: "none running", out: `{"State":"exited"}` + "\n", want: "stopped"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusOf([]byte(tt.out)); got != tt.want {
				t.Errorf("statusOf(%q) = %q, want %q", tt.out, got, tt.want)
			}
		})
	}
}
