package contractcheck

import (
	"context"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// buf runs the pinned buf tool from the proto directory the test has changed into.
func buf(t *testing.T, timeout time.Duration, args ...string) ([]byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), timeout)
	defer cancel()
	command := append([]string{"tool", "-modfile=../tools/go.mod", "buf"}, args...)
	return exec.CommandContext(ctx, "go", command...).CombinedOutput()
}

// sources reads every file under root, keyed by its slash-separated path, with line endings
// normalised so a Windows checkout compares equal to the generator's output.
func sources(t *testing.T, root string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)] = strings.ReplaceAll(string(content), "\r\n", "\n")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func write(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
}
