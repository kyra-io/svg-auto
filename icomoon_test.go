package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestIsZipFile(t *testing.T) {
	zipPath := writeTestZip(t, map[string]string{"file.txt": "content"})
	if !isZipFile(zipPath) {
		t.Fatal("expected a valid zip file")
	}

	notZip := filepath.Join(t.TempDir(), "not-a-zip")
	if err := os.WriteFile(notZip, []byte("plain text"), 0o644); err != nil {
		t.Fatal(err)
	}
	if isZipFile(notZip) {
		t.Fatal("plain text must not be detected as a zip")
	}
	if isZipFile(filepath.Join(t.TempDir(), "missing")) {
		t.Fatal("a missing file must not be detected as a zip")
	}
}

func TestWaitForFileStable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "download")
	if err := os.WriteFile(path, []byte("complete"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := waitForFileStable(ctx, path, 2*time.Second); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWaitForFileStableCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := waitForFileStable(ctx, filepath.Join(t.TempDir(), "missing"), time.Second)
	if err == nil || !strings.Contains(err.Error(), "did not finish downloading") {
		t.Fatalf("expected a cancellation error, got %v", err)
	}
}
