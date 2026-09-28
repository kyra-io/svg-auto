//go:build e2e

package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEndToEndIcoMoon(t *testing.T) {
	root := t.TempDir()
	sourceDir := filepath.Join(root, "source")
	projectDir := filepath.Join(root, "project")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}

	sourceSVG := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"><path d="M3 3h18v18H3z"/></svg>`
	writeE2EFile(t, filepath.Join(sourceDir, "ci-test.svg"), sourceSVG)
	writeE2EFile(t, filepath.Join(projectDir, "sprite.svg"), "<svg>\n<defs>\n</defs>\n</svg>\n")
	writeE2EFile(t, filepath.Join(projectDir, "style.css"), ".icon {}\n")
	writeE2EFile(t, filepath.Join(projectDir, "selection.json"), `{
  "metadata": {"name": "CI"},
  "iconSets": [{"selection": [], "id": 0, "metadata": {}, "height": 1024, "prevSize": 32, "icons": []}],
  "uid": -1,
  "preferences": {}
}`)

	cfg := Config{
		ProjectPath: projectDir,
		IconPrefix:  "e2e-",
		Files: []FileRule{
			{Name: "sprite.svg", Mode: "text", Marker: "</defs>", Position: "before", Template: `<symbol id="{{.Prefix}}{{.Name}}" viewBox="{{.ViewBox}}">{{.Body}}</symbol>`},
			{Name: "selection.json", Mode: "icomoon"},
			{Name: "style.css", Mode: "text", Position: "end", Template: `.{{.Prefix}}{{.Name}} { width: 1em; }`},
		},
	}
	configData, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "config.json")
	if err := os.WriteFile(configPath, configData, 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("SVG_AUTO_CONFIG", configPath)
	t.Chdir(sourceDir)
	originalArgs := os.Args
	os.Args = []string{"svg-auto", "."}
	t.Cleanup(func() { os.Args = originalArgs })

	if err := run(); err != nil {
		t.Fatalf("end-to-end run failed: %v", err)
	}

	assertE2EContains(t, filepath.Join(projectDir, "sprite.svg"), `id="e2e-ci-test"`)
	assertE2EContains(t, filepath.Join(projectDir, "style.css"), `.e2e-ci-test`)
	assertE2EContains(t, filepath.Join(projectDir, "selection.json"), `"ci-test"`)

	zips, err := filepath.Glob(filepath.Join(sourceDir, outputDir, "*.zip"))
	if err != nil {
		t.Fatal(err)
	}
	if len(zips) != 1 || !isZipFile(zips[0]) {
		t.Fatalf("expected one downloaded zip, got %v", zips)
	}
}

func writeE2EFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertE2EContains(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), want) {
		t.Fatalf("%s does not contain %q:\n%s", path, want, data)
	}
}
