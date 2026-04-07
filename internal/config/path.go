package config

import (
	"os"
	"path/filepath"
)

// ResolvePath turns project-relative paths into absolute ones.
func ResolvePath(path string) string {
	if path == "" || path == ":memory:" || filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(appRoot(), path)
}

// appRoot tries to find the project root so files work from different run locations.
func appRoot() string {
	if root, ok := findProjectRootFrom(mustGetwd()); ok {
		return root
	}

	if exe, err := os.Executable(); err == nil {
		if root, ok := findProjectRootFrom(filepath.Dir(exe)); ok {
			return root
		}

		return filepath.Dir(exe)
	}

	return mustGetwd()
}

// findProjectRootFrom walks up folders until it finds the project marker file.
func findProjectRootFrom(start string) (string, bool) {
	dir := start
	for {
		if fileExists(filepath.Join(dir, "schema.sql")) {
			return dir, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

// fileExists checks that a path exists and is not a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// mustGetwd returns the current folder, with "." as a safe fallback.
func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
