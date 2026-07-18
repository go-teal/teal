package generators

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-teal/teal/pkg/configs"
)

// The Makefile is developer-editable scaffolding; teal gen must not clobber
// an existing one. Mirrors the skip-if-exists behaviour of go.mod / main.go /
// Dockerfile generators.
func TestGenMakefileSkipsExistingFile(t *testing.T) {
	dir := t.TempDir()
	cfg := &configs.Config{ProjectPath: dir}
	profile := &configs.ProjectProfile{}
	g := InitGenMakefile(cfg, profile)

	// Pre-existing hand-edited Makefile must survive gen untouched.
	custom := []byte("# custom\nui:\n\techo hi\n")
	if err := os.WriteFile(filepath.Join(dir, "Makefile"), custom, 0o644); err != nil {
		t.Fatal(err)
	}

	err, skipped := g.RenderToFile()
	if err != nil {
		t.Fatalf("RenderToFile: %v", err)
	}
	if !skipped {
		t.Fatal("expected existing Makefile to be skipped, was overwritten")
	}
	got, err := os.ReadFile(filepath.Join(dir, "Makefile"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(custom) {
		t.Fatalf("existing Makefile was modified:\n%s", got)
	}
}

func TestGenMakefileGeneratesWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	cfg := &configs.Config{ProjectPath: dir}
	profile := &configs.ProjectProfile{Name: "demo"}
	g := InitGenMakefile(cfg, profile)

	err, skipped := g.RenderToFile()
	if err != nil {
		t.Fatalf("RenderToFile: %v", err)
	}
	if skipped {
		t.Fatal("expected a fresh Makefile to be generated, got skipped")
	}
	if _, err := os.Stat(filepath.Join(dir, "Makefile")); err != nil {
		t.Fatalf("Makefile not written: %v", err)
	}
}
