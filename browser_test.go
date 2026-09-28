package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindBrowserFromOverride(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv(browserEnvVar, executable)

	got, err := findBrowser()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Clean(got) != filepath.Clean(executable) {
		t.Errorf("expected %q, got %q", executable, got)
	}
}

func TestFindBrowserInvalidOverride(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-browser")
	t.Setenv(browserEnvVar, missing)

	_, err := findBrowser()
	if err == nil || !strings.Contains(err.Error(), browserEnvVar) {
		t.Fatalf("expected an override error, got %v", err)
	}
}

func TestAbsoluteBrowserPathsAreNotEmpty(t *testing.T) {
	paths := absBrowserPaths()
	if len(paths) == 0 {
		t.Fatal("expected browser paths for the current operating system")
	}
	for _, path := range paths {
		if path == "" {
			t.Error("browser path must not be empty")
		}
	}
}
