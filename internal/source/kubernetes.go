package source

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/weaveworks/weave-policy-validator/internal/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Kubernetes struct {
	Path string
}

func NewKubernetesSource(path string) *Kubernetes {
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
		fileInfo, err := ioutil.ReadDir(k.Path)
		if err != nil {
			return false
		}
		for _, file := range fileInfo {
			if isValidKubeFile(filepath.Join(k.Path, file.Name())) {
				return true
			}
		}
		return false
	}

	return isYamlFile(k.Path)
}

func glob(path string) ([]string, error) {
	var paths []string
	err := filepath.Walk(path, func(path string, _ os.FileInfo, err error) error {
		if isHiddenFile(path) {
			return nil
		}
		if isYamlFile(path) {
			paths = append(paths, path)
		}
		return err
	})
	return paths, err
}

func isYamlFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".yml" || ext == ".yaml"
}

func isValidKubeFile(path string) bool {
	if !isYamlFile(path) {
		return false
	}

	if isHiddenFile(path) {
		return false
	}

	node, err := yaml.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return false
	}

	if node.GetKind() == "" {
		return false
	}
	return true
}
