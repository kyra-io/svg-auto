package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeArgTestFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("<svg></svg>"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveSVGArgsCurrentDirectory(t *testing.T) {
	dir := t.TempDir()
	writeArgTestFile(t, filepath.Join(dir, "a.svg"))
	writeArgTestFile(t, filepath.Join(dir, "b.SVG"))
	writeArgTestFile(t, filepath.Join(dir, "notes.txt"))
	if err := os.Mkdir(filepath.Join(dir, "nested.svg"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	files, err := resolveSVGArgs([]string{"."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 SVG files, got %d: %v", len(files), files)
	}
	if filepath.Base(files[0]) != "a.svg" || filepath.Base(files[1]) != "b.SVG" {
		t.Errorf("unexpected files or order: %v", files)
	}
	for _, file := range files {
		if !filepath.IsAbs(file) {
			t.Errorf("expected an absolute path, got %q", file)
		}
	}
}

func TestResolveSVGArgsCurrentDirectoryWithExplicitFile(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.svg")
	b := filepath.Join(dir, "b.svg")
	writeArgTestFile(t, a)
	writeArgTestFile(t, b)
	t.Chdir(dir)

	files, err := resolveSVGArgs([]string{"a.svg", "."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected duplicate paths to be removed, got %d: %v", len(files), files)
	}
	if files[0] != a || files[1] != b {
		t.Errorf("unexpected files or order: %v", files)
	}
}

func TestResolveSVGArgsCurrentDirectoryWithoutSVGs(t *testing.T) {
	dir := t.TempDir()
	writeArgTestFile(t, filepath.Join(dir, "notes.txt"))
	t.Chdir(dir)

	_, err := resolveSVGArgs([]string{"."})
	if err == nil || !strings.Contains(err.Error(), "no SVG files") {
		t.Fatalf("expected a no SVG files error, got %v", err)
	}
}

func TestResolveSVGArgsWithoutArguments(t *testing.T) {
	_, err := resolveSVGArgs(nil)
	if err == nil || !strings.Contains(err.Error(), "Usage:") {
		t.Fatalf("expected usage error, got %v", err)
	}
}
