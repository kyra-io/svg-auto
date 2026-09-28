package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const usage = `Usage: svg-auto <file.svg|.> [more.svg ...]

Imports SVG files into IcoMoon and downloads the generated package (.zip) to ./output/.
Use . to import every SVG file in the current directory.

Options:
  -h, --help    show this help

Environment variables:
  SVG_AUTO_BROWSER    path or name of the browser executable (optional)`

func parseArgs() ([]string, error) {
	args := os.Args[1:]

	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Println(usage)
			os.Exit(0)
		}
	}

	return resolveSVGArgs(args)
}

func resolveSVGArgs(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("%s", usage)
	}

	files := make([]string, 0, len(args))
	seen := make(map[string]bool)
	add := func(path string) error {
		if err := validateSVG(path); err != nil {
			return err
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get the absolute path of %q: %w", path, err)
		}
		if !seen[abs] {
			seen[abs] = true
			files = append(files, abs)
		}
		return nil
	}

	for _, arg := range args {
		if arg == "." {
			entries, err := os.ReadDir(".")
			if err != nil {
				return nil, fmt.Errorf("failed to read the current directory: %w", err)
			}

			found := 0
			for _, entry := range entries {
				if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".svg") {
					continue
				}
				info, err := entry.Info()
				if err != nil {
					return nil, fmt.Errorf("failed to inspect %q: %w", entry.Name(), err)
				}
				if !info.Mode().IsRegular() {
					continue
				}
				if err := add(entry.Name()); err != nil {
					return nil, err
				}
				found++
			}
			if found == 0 {
				return nil, fmt.Errorf("no SVG files found in the current directory")
			}
			continue
		}

		if err := add(arg); err != nil {
			return nil, err
		}
	}
	return files, nil
}

func validateSVG(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %q", path)
		}
		return fmt.Errorf("failed to access %q: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("expected a file, but %q is a directory", path)
	}
	if !strings.EqualFold(filepath.Ext(path), ".svg") {
		return fmt.Errorf("%q does not end in .svg", path)
	}
	return nil
}
