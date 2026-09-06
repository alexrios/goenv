package commands

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReloadEnvJSONValidation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fake Go command uses a POSIX shell")
	}
	old := goPath
	t.Cleanup(func() { goPath = old })
	for _, tc := range []struct {
		name, all, changed string
		wantErr            bool
	}{
		{"valid", `{"GOPATH":"/go","GOPROXY":"<proxy>&"}`, `{"GOPATH":"/go"}`, false},
		{"duplicate", `{"GOPATH":"/a","GOPATH":"/b"}`, `{}`, true},
		{"invalid UTF-8", "{\"GOPATH\":\"\xff\"}", `{}`, true},
		{"duplicate changed", `{"GOPATH":"/go"}`, `{"GOPATH":"/a","GOPATH":"/b"}`, true},
		{"invalid changed UTF-8", `{"GOPATH":"/go"}`, "{\"GOPATH\":\"\xff\"}", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, data := range map[string]string{"all": tc.all, "changed": tc.changed, "go": "#!/bin/sh\ncd -- \"$(dirname -- \"$0\")\"\nif [ \"$2\" = -changed ]; then cat changed; else cat all; fi\n"} {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0755); err != nil {
					t.Fatal(err)
				}
			}
			goPath = filepath.Join(dir, "go")
			items, err := ReloadEnv()
			if (err != nil) != tc.wantErr {
				t.Fatalf("items=%v err=%v", items, err)
			}
			if !tc.wantErr {
				if len(items) != 2 {
					t.Fatal(items)
				}
				for _, v := range items {
					if v.Key == "GOPATH" && (!v.Changed || v.Value != "/go") {
						t.Fatal(v)
					}
				}
			}
		})
	}
}
