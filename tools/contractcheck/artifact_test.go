package contractcheck

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
)

const modulePath = "github.com/schemo-app/proto"

// TestModuleZipConsumer installs a Go module artifact without a workspace or local replace.
func TestModuleZipConsumer(t *testing.T) {
	sourceRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	contract := module.Version{Path: modulePath, Version: copyModuleSources(t, sourceRoot, root)}
	proxyURL := serveModuleProxy(t, contract, root)
	consumer := writeConsumer(t, contract, root)
	runConsumerTests(t, consumer, contract, proxyURL)
}

// copyModuleSources copies the publishable files of sourceRoot into root and returns a version
// named after their content. A content-addressed prerelease avoids rewriting an immutable version
// in the Go cache.
func copyModuleSources(t *testing.T, sourceRoot, root string) string {
	t.Helper()
	digest := sha256.New()
	err := filepath.WalkDir(sourceRoot, func(path string, entry fs.DirEntry, err error) error {
		if err == nil && entry.IsDir() && isUnpublished(entry.Name()) {
			return fs.SkipDir
		}
		if err != nil || entry.IsDir() {
			return err
		}
		relative, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		target := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, content, 0o644); err != nil {
			return err
		}
		fmt.Fprintf(digest, "%s\x00%d\x00", filepath.ToSlash(relative), len(content))
		_, err = digest.Write(content)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("v0.0.0-contracttest.%x", digest.Sum(nil)[:12])
}

// isUnpublished reports whether a directory is left out of the module zip.
func isUnpublished(name string) bool {
	return name == "node_modules" || name == "tools" || name == "tmp" || strings.HasPrefix(name, ".")
}

// serveModuleProxy lays out a file-based GOPROXY holding the module zip built from root and
// returns its URL.
func serveModuleProxy(t *testing.T, contract module.Version, root string) string {
	t.Helper()
	proxy := t.TempDir()
	versions := filepath.Join(proxy, filepath.FromSlash(contract.Path), "@v")
	if err := os.MkdirAll(versions, 0o755); err != nil {
		t.Fatal(err)
	}
	archivePath := filepath.Join(versions, contract.Version+".zip")
	archive, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	createErr := zip.CreateFromDir(archive, contract, root)
	closeErr := archive.Close()
	if createErr != nil || closeErr != nil {
		t.Fatalf("create module zip: %v; close: %v", createErr, closeErr)
	}
	if _, err := zip.CheckZip(contract, archivePath); err != nil {
		t.Fatal(err)
	}
	manifest, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(versions, contract.Version+".mod"), manifest)
	write(t, filepath.Join(versions, contract.Version+".info"), fmt.Appendf(nil,
		`{"Version":%q,"Time":"2026-10-13T00:00:00Z"}`, contract.Version))
	write(t, filepath.Join(versions, "list"), []byte(contract.Version+"\n"))

	proxyPath := filepath.ToSlash(proxy)
	if !strings.HasPrefix(proxyPath, "/") {
		proxyPath = "/" + proxyPath
	}
	return (&url.URL{Scheme: "file", Path: proxyPath}).String()
}

// writeConsumer creates a standalone module that requires the contract and carries the transport
// test, so the test runs against the published artifact and not the working tree.
func writeConsumer(t *testing.T, contract module.Version, root string) string {
	t.Helper()
	consumer := t.TempDir()
	write(t, filepath.Join(consumer, "go.mod"), fmt.Appendf(nil,
		"module example.test/contract-consumer\n\ngo 1.26.7\n\nrequire %s %s\n", contract.Path, contract.Version))
	for _, name := range []string{"transport_test.go", "go.sum"} {
		content, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		write(t, filepath.Join(consumer, name), content)
	}
	return consumer
}

// runConsumerTests runs the consumer's tests with only the file proxy able to supply the contract.
func runConsumerTests(t *testing.T, consumer string, contract module.Version, proxyURL string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "-mod=mod", "./...")
	command.Dir = consumer
	command.Env = append(os.Environ(), "GOWORK=off", "GOPRIVATE=", "GONOPROXY=",
		"GONOSUMDB="+contract.Path, "GOPROXY="+proxyURL+",https://proxy.golang.org")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("standalone consumer: %v\n%s", err, output)
	}
}
