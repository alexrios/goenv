package goenv

import (
	"encoding/json/v2"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestExperimentVersionBoundaries(t *testing.T) {
	for _, tc := range []struct {
		version         string
		old, json, simd bool
	}{
		{"go1.24.9", true, false, false}, {"go1.25.0", true, true, false}, {"go1.25.9", true, true, false}, {"go1.26.0", false, true, true}, {"go1.27.0", false, true, true}, {"unknown", true, true, true},
	} {
		t.Run(tc.version, func(t *testing.T) {
			got := FilterKnownValuesForVersion("GOEXPERIMENT", ParseGoVersion(tc.version))
			for key, want := range map[string]bool{"cacheprog": tc.old, "rangefunc": tc.old, "jsonv2": tc.json, "nojsonv2": tc.json, "simd": tc.simd} {
				if slices.Contains(got, key) != want {
					t.Errorf("%s presence: %v, want %v", key, got, want)
				}
			}
			if tc.version == "unknown" && !slices.Equal(got, GetEnvVarKnownValues("GOEXPERIMENT")) {
				t.Fatal("unknown version changed")
			}
		})
	}
}

func TestJSONExportCompatibility(t *testing.T) {
	for _, vars := range [][]EnvVar{nil, {}, {{Key: "GOPROXY", Value: "<proxy>&"}, {Key: "GOPATH", Value: "/go"}}} {
		output, err := GenerateShellExport(vars, ShellJSON, ExportAll)
		if err != nil {
			t.Fatal(err)
		}
		want := map[string]string{}
		for _, v := range vars {
			want[v.Key] = v.Value
		}
		var got map[string]string
		if err := json.Unmarshal([]byte(output), &got); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v want %v", got, want)
		}
		if strings.Contains(output, `\u0026`) || strings.Contains(output, `\u003c`) {
			t.Fatalf("HTML escaping: %s", output)
		}
	}
	for _, v := range []EnvVar{{Key: "GOPATH", Value: "\xff"}, {Key: "\xff", Value: "/go"}} {
		output, err := GenerateShellExport([]EnvVar{v}, ShellJSON, ExportAll)
		if err == nil || output != "" {
			t.Fatalf("invalid UTF-8: output=%q err=%v", output, err)
		}
	}
}
