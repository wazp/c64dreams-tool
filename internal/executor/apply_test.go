package executor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wazp/c64dreams-tool/internal/layout"
	"github.com/wazp/c64dreams-tool/pkg/model"
)

func TestApplyDryRun(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "input")
	dst := filepath.Join(root, "output")

	mustMkdir(t, src)
	mustWrite(t, filepath.Join(src, "game/disk1.d64"), []byte("data"))

	planned := []layout.PlannedFile{{
		GameID: "g1",
		Path:   "game/disk1.d64",
		Target: model.TargetSD2IEC,
	}}

	results, err := Apply(planned, Options{InputRoot: src, OutputRoot: dst, DryRun: true})
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if len(results) != 2 { // mkdir + copy intent
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if _, err := os.Stat(filepath.Join(dst, "game", "disk1.d64")); err == nil {
		t.Fatalf("dry-run should not create files")
	}
}

func TestApplyRealCopy(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "input")
	dst := filepath.Join(root, "output")

	mustMkdir(t, src)
	mustWrite(t, filepath.Join(src, "game/disk1.d64"), []byte("data"))

	planned := []layout.PlannedFile{{
		GameID: "g1",
		Path:   "game/disk1.d64",
		Target: model.TargetSD2IEC,
	}}

	results, err := Apply(planned, Options{InputRoot: src, OutputRoot: dst, DryRun: false})
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	data, err := os.ReadFile(filepath.Join(dst, "game", "disk1.d64"))
	if err != nil {
		t.Fatalf("dest missing: %v", err)
	}
	if string(data) != "data" {
		t.Fatalf("unexpected dest content: %s", string(data))
	}
}

func TestApplyOverwriteProtection(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "input")
	dst := filepath.Join(root, "output")

	mustMkdir(t, src)
	mustMkdir(t, filepath.Join(dst, "game"))
	mustWrite(t, filepath.Join(src, "game/disk1.d64"), []byte("source"))
	mustWrite(t, filepath.Join(dst, "game/disk1.d64"), []byte("existing"))

	planned := []layout.PlannedFile{{
		GameID: "g1",
		Path:   "game/disk1.d64",
		Target: model.TargetSD2IEC,
	}}

	results, err := Apply(planned, Options{InputRoot: src, OutputRoot: dst, Overwrite: false})
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if results[0].Action != "skip" {
		t.Fatalf("expected skip when not overwriting, got %s", results[0].Action)
	}

	data, _ := os.ReadFile(filepath.Join(dst, "game/disk1.d64"))
	if string(data) != "existing" {
		t.Fatalf("existing file should remain untouched")
	}

	_, err = Apply(planned, Options{InputRoot: src, OutputRoot: dst, Overwrite: true})
	if err != nil {
		t.Fatalf("Apply returned error on overwrite: %v", err)
	}

	data, _ = os.ReadFile(filepath.Join(dst, "game/disk1.d64"))
	if string(data) != "source" {
		t.Fatalf("expected file to be overwritten")
	}
}

func TestApplyDeterministicOrder(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "input")
	dst := filepath.Join(root, "output")

	mustMkdir(t, src)
	mustWrite(t, filepath.Join(src, "b/b.d64"), []byte("b"))
	mustWrite(t, filepath.Join(src, "a/a.d64"), []byte("a"))

	planned := []layout.PlannedFile{
		{GameID: "b", Path: "b/b.d64", Target: model.TargetSD2IEC},
		{GameID: "a", Path: "a/a.d64", Target: model.TargetSD2IEC},
	}

	results, err := Apply(planned, Options{InputRoot: src, OutputRoot: dst, DryRun: true})
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if results[0].Dest > results[1].Dest {
		t.Fatalf("expected deterministic sorted order")
	}
}

func mustMkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
}

func mustWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir for file: %v", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
