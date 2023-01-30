package kustomization

import (
	"context"
	"os"
	"path/filepath"

	"github.com/MagalixTechnologies/weave-iac-validator/internal/types"
)

type Kubernetes struct {
	Path string
}

func NewKubernetesKustomizer(path string) *Kubernetes {
	return &Kubernetes{Path: path}
}

func (k *Kubernetes) Type() string {
	return KubernetesType
}

func (k *Kubernetes) ResourceFiles(_ context.Context) ([]*types.File, error) {
	paths, err := glob(k.Path)
	if err != nil {
		return nil, err
	}
	var files []*types.File
	for _, path := range paths {
		if !isYamlFile(path) {
			continue
		}
		file, err := types.NewFileFromPath(path)
		if err != nil {
			return nil, err
		}
		for _, resource := range file.Resources {
			resource.Rendered = resource.Raw
		}
		files = append(files, file)
	}
	return files, nil
}

func (k *Kubernetes) IsValidPath() bool {
	info, err := os.Stat(k.Path)
	if err != nil {
		return false
	}

	if info.IsDir() {
		return true
	}

	return isYamlFile(k.Path)
}

func isYamlFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".yml" || ext == ".yaml"
}
