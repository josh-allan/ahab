package ahab

import (
	"reflect"
	"testing"
)

func Test_readIgnoreFile(t *testing.T) {
	type args struct {
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]struct{}
		wantErr bool
	}{
		{
			name: "Test with valid directory",
			args: args{dir: "./testdata"},
			want: map[string]struct{}{
				"test1.yaml": {},
				"test2.yaml": {},
			},
			wantErr: false,
		},
		{
			name:    "Test with non-existent directory",
			args:    args{dir: "./non_existent_dir"},
			want:    make(map[string]struct{}),
			wantErr: false,
		},
		{
			name:    "Test with empty directory",
			args:    args{dir: "./empty_dir"},
			want:    make(map[string]struct{}),
			wantErr: false,
		},
		{
			name:    "empty directory",
			args:    args{dir: "testdata/empty"},
			want:    map[string]struct{}{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readIgnoreFile(tt.args.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("readIgnoreFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readIgnoreFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getDockerFiles(t *testing.T) {
	type args struct {
		dir string
	}
	var tests []struct {
		name string
		args args
		want []string
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDockerFiles(tt.args.dir); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDockerFiles() = %v, want %v", got, tt.want)
			}
		})
	}
}
