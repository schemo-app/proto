package contractcheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestWireBaselineRejectsFieldRenumbering verifies that the compatibility gate is effective.
func TestWireBaselineRejectsFieldRenumbering(t *testing.T) {
	t.Chdir("../../proto")
	fresh := t.TempDir()
	if err := os.CopyFS(fresh, os.DirFS(".")); err != nil {
		t.Fatal(err)
	}
	baseline, err := filepath.Abs("compatibility/v1.binpb")
	if err != nil {
		t.Fatal(err)
	}
	config, err := filepath.Abs("buf.wire.yaml")
	if err != nil {
		t.Fatal(err)
	}
	checkWire := func() ([]byte, error) {
		return buf(t, 3*time.Minute, "breaking", fresh,
			"--config", config, "--against", baseline, "--error-format", "json")
	}
	if output, err := checkWire(); err != nil {
		t.Fatalf("unchanged wire failed: %v\n%s", err, output)
	}

	path := filepath.Join(fresh, "platform", "v1", "platform.proto")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	renumbered := strings.Replace(string(source), "string name = 1;", "string name = 101;", 1)
	if renumbered == string(source) {
		t.Fatal("fixture did not change the field number")
	}
	write(t, path, []byte(renumbered))
	output, err := checkWire()
	if err == nil || !strings.Contains(string(output), "FIELD_NO_DELETE_UNLESS_NUMBER_RESERVED") {
		t.Fatalf("renumbered field was not rejected by the wire rule: %v\n%s", err, output)
	}
}
