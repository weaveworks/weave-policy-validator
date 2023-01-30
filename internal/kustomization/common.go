package kustomization

import (
	"os"
	"path/filepath"
)

func glob(path string) ([]string, error) {
	var paths []string
	err := filepath.Walk(path, func(path string, _ os.FileInfo, err error) error {
		paths = append(paths, path)
		return err
	})
	return paths, err
}
