package kustomization

import (
	"path/filepath"
	"strings"
)

func isHiddenFile(path string) bool {
	return strings.HasPrefix(filepath.Base(path), ".")
}
