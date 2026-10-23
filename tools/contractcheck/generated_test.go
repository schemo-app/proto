package contractcheck

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// TestGeneratedGoContracts checks this repository, even when no desktop checkout exists.
func TestGeneratedGoContracts(t *testing.T) {
	requireGeneratedInSync(t, "", "go")
}

// TestGeneratedESContracts checks the ECMAScript modules the npm package ships.
func TestGeneratedESContracts(t *testing.T) {
	requireGeneratedInSync(t, "buf.gen.es.yaml", "typescript")
}

// requireGeneratedInSync generates into a temporary directory and compares the result with the
// committed output directory. An empty template selects the default buf.gen.yaml.
func requireGeneratedInSync(t *testing.T, template, output string) {
	t.Helper()
	t.Chdir("../../proto")
	fresh := t.TempDir()
	args := []string{"generate", "--output", fresh}
	if template != "" {
		args = append(args, "--template", template)
	}
	if message, err := buf(t, 2*time.Minute, args...); err != nil {
		t.Fatalf("generate %s contracts: %v\n%s", output, err, message)
	}
	generated := sources(t, filepath.Join(fresh, output))
	if len(generated) == 0 {
		t.Fatalf("generation produced no %s sources", output)
	}
	if !reflect.DeepEqual(generated, sources(t, filepath.Join("..", output))) {
		t.Fatalf("generated %s contracts differ from proto; run task generate", output)
	}
}
