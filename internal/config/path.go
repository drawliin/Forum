package config

import (
	"os"
	"path/filepath"
)

func ResolvePath(path string) string {
	if path == "" || path == ":memory:" || filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(appRoot(), path)
}

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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
