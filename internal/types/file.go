package types

import (
	"fmt"

	"github.com/weaveworks/weave-policy-validator/internal/yaml"
)

type File struct {
	Path       string
	Remediated bool
	Resources  map[string]*Resource
}

// NewFile creates new empty file
func NewFile(path string) *File {
	return &File{
		Path:      path,
		Resources: make(map[string]*Resource),
	}
}

// NewFileFromPath creates new file from given path
func NewFileFromPath(path string) (*File, error) {
	nodes, err := yaml.MultiDocFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file, path: %s error: %v", path, err)
	}

	resources := make(map[string]*Resource)
	for i := range nodes {
		obj := NewObject(nodes[i])
		resources[obj.ID()] = &Resource{
			Raw: obj,
		}
	}

	return &File{
		Path:      path,
		Resources: resources,
	}, nil
}

// ResourceExists checks if a resource exists in the file
func (f *File) ResourceExists(id string) bool {
	_, found := f.Resources[id]
	return found
}

// SetResourceRenderedObject sets rendered object
func (f *File) SetResourceRenderedObject(obj *Object) {
	f.Resources[obj.ID()].Rendered = obj
}

// Content returns file content in string format
func (f *File) Content() (string, error) {
	var nodes []*yaml.Node
	for _, resource := range f.Resources {
		if resource.Raw != nil {
			nodes = append(nodes, resource.Raw.node)
		}
	}
	raw, err := yaml.Bytes(nodes)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
