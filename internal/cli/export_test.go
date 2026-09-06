package cli

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/alexrios/goenv/internal/goenv"
	"github.com/alexrios/goenv/internal/testutil"
)

func TestCLIExportSerializationError(t *testing.T) {
	if os.Getenv("GOENV_TEST_INVALID_EXPORT") == "1" {
		printCLIExport([]goenv.EnvVar{{Key: "GOPATH", Value: "\xff"}}, goenv.ShellJSON, goenv.ExportAll)
		return
	}
	testutil.TempConfig(t)
	cmd := exec.Command(os.Args[0], "-test.run=^TestCLIExportSerializationError$")
	cmd.Env = append(os.Environ(), "GOENV_TEST_INVALID_EXPORT=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 {
		t.Fatalf("want exit 1, got %v", err)
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "Error: failed to marshal JSON export") {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}
