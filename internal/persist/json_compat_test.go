package persist

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestJSONCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name, legacy, wrongCase, duplicate, invalid, collection string
		save                                                    func(string) error
		load                                                    func(string) (any, error)
		want                                                    any
	}{
		{"config", `{"sortMode":"alpha","favorites":null,"future":true}`, `{"SortMode":"category"}`, `{"sortMode":"alpha","sortMode":"category"}`, "{\"theme\":\"\xff\"}", "favorites",
			func(string) error { return SaveConfig(Config{SortMode: "alpha"}) },
			func(string) (any, error) { return LoadConfig() }, Config{SortMode: "alpha"}},
		{"history", `{"records":null,"currentIdx":0,"future":true}`, `{"Records":[{"key":"GOPATH"}]}`, `{"records":[],"records":[]}`, "{\"records\":[{\"key\":\"\xff\"}]}", "records",
			func(string) error { return SaveHistory(SessionHistory{}, 50) },
			func(string) (any, error) { return LoadHistory() }, SessionHistory{}},
		{"snapshot", `{"name":"old","createdAt":"0001-01-01T00:00:00Z","variables":null,"future":true}`, `{"Name":"wrong"}`, `{"variables":{"GOPATH":"a","GOPATH":"b"}}`, "{\"variables\":{\"GOPATH\":\"\xff\"}}", "variables",
			func(path string) error { return ExportSnapshot(Snapshot{Name: "old"}, path) },
			func(path string) (any, error) { return ImportSnapshot(path) }, Snapshot{Name: "old"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			dir, err := os.UserConfigDir()
			if err != nil {
				t.Fatal(err)
			}
			dir = filepath.Join(dir, ConfigDirName)
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(dir, tc.name+".json")
			write := func(s string) {
				t.Helper()
				if err := os.WriteFile(path, []byte(s), 0644); err != nil {
					t.Fatal(err)
				}
			}
			write(tc.legacy)
			got, err := tc.load(path)
			if err != nil || !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("legacy: got %#v, %v; want %#v", got, err, tc.want)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != tc.legacy {
				t.Fatal("load rewrote legacy file")
			}
			write(tc.wrongCase)
			got, err = tc.load(path)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.ValueOf(got).IsZero() {
				t.Fatalf("wrong case populated fields: %#v", got)
			}
			for _, bad := range []string{tc.duplicate, tc.invalid} {
				write(bad)
				if _, err := tc.load(path); err == nil {
					t.Fatalf("accepted malformed JSON: %q", bad)
				}
			}
			if err := tc.save(path); err != nil {
				t.Fatal(err)
			}
			data, err = os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var fields map[string]jsontext.Value
			if err := json.Unmarshal(data, &fields); err != nil {
				t.Fatal(err)
			}
			wantCollection := "[]"
			if tc.name == "snapshot" {
				wantCollection = "{}"
			}
			if tc.name == "config" {
				wantCollection = ""
			}
			if string(fields[tc.collection]) != wantCollection {
				t.Fatalf("nil collection: %s", data)
			}
			got, err = tc.load(path)
			if err != nil {
				t.Fatal(err)
			}
			// v2 deliberately turns nil collections into non-nil empty collections.
			switch v := got.(type) {
			case SessionHistory:
				if len(v.Records) != 0 {
					t.Fatal(v)
				}
			case Snapshot:
				if len(v.Variables) != 0 || v.Name != "old" {
					t.Fatal(v)
				}
			case Config:
				if v.SortMode != "alpha" {
					t.Fatal(v)
				}
			}
		})
	}
}

func TestConfigJSONOmission(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	for _, favorites := range [][]string{nil, {}, {"GOPATH"}} {
		cfg := Config{SortMode: "alpha", Favorites: favorites}
		if err := SaveConfig(cfg); err != nil {
			t.Fatal(err)
		}
		path, err := configPath()
		if err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]jsontext.Value
		if err := json.Unmarshal(data, &fields); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"watchInterval", "maxHistory"} {
			if _, ok := fields[key]; ok {
				t.Fatalf("zero %s present: %s", key, data)
			}
		}
		_, present := fields["favorites"]
		if present != (favorites != nil) {
			t.Fatalf("favorites omission: %s", data)
		}
		got, err := LoadConfig()
		if err != nil || !reflect.DeepEqual(got, cfg) {
			t.Fatalf("round trip: %#v, %v", got, err)
		}
	}
}

func TestSerializationFailurePreservesFiles(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfgPath, err := configPath()
	if err != nil {
		t.Fatal(err)
	}
	histPath, err := historyPath()
	if err != nil {
		t.Fatal(err)
	}
	presetDir, err := PresetsDir()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, path string
		save       func(string) error
	}{
		{"config", cfgPath, func(string) error { return SaveConfig(Config{Theme: "\xff"}) }},
		{"history", histPath, func(string) error { return SaveHistory(SessionHistory{Records: []EditRecord{{Key: "\xff"}}}, 50) }},
		{"snapshot", filepath.Join(t.TempDir(), "snapshot.json"), func(path string) error {
			return ExportSnapshot(Snapshot{Variables: map[string]string{"GOPATH": "\xff"}}, path)
		}},
		{"preset", filepath.Join(presetDir, "preset.json"), func(path string) error { return ExportSnapshot(Snapshot{Name: "\xff"}, path) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const original = `{"keep":"existing file"}`
			if err := os.WriteFile(tc.path, []byte(original), 0644); err != nil {
				t.Fatal(err)
			}
			if err := tc.save(tc.path); err == nil {
				t.Fatal("expected serialization failure")
			}
			data, err := os.ReadFile(tc.path)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != original {
				t.Fatalf("file changed: %q", data)
			}
		})
	}
}

func TestLegacyPreset(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	dir, err := PresetsDir()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "old.json")
	data := []byte(`{"name":"old","createdAt":"2026-01-01T00:00:00Z","variables":{"GOPATH":"/go","GOPROXY":"https://example.test?a=1\u0026b=2"}}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	presets, skipped, err := ListPresets()
	if err != nil || skipped != 0 || len(presets) != 1 {
		t.Fatalf("presets=%v skipped=%d err=%v", presets, skipped, err)
	}
	if presets[0].Variables["GOPROXY"] != "https://example.test?a=1&b=2" {
		t.Fatal(presets)
	}
	if err := ExportSnapshot(presets[0].Snapshot, path); err != nil {
		t.Fatal(err)
	}
	got, err := ImportSnapshot(path)
	if err != nil || !reflect.DeepEqual(got, presets[0].Snapshot) {
		t.Fatalf("round trip: %#v %v", got, err)
	}
}
