package kustomization

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/MagalixTechnologies/weave-iac-validator/internal/types"
	"github.com/MagalixTechnologies/weave-iac-validator/internal/yaml"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/engine"
)

type Helm struct {
	Path      string
	valueFile *string
}

func NewHelmKustomizer(path string) *Helm {
	return &Helm{Path: path}
}

func (h *Helm) Type() string {
	return HelmType
}

func (h *Helm) SetValueFile(filename string) {
	h.valueFile = &filename
}

func (h *Helm) ResourceFiles(_ context.Context) ([]*types.File, error) {
	chart, err := loader.Load(h.Path)
	if err != nil {
		return nil, err
	}

	vals := chart.Values
	if h.valueFile != nil {
		values, err := chartutil.ReadValuesFile(*h.valueFile)
		if err != nil {
			return nil, err
		}
		vals, err = chartutil.CoalesceValues(chart, values)
		if err != nil {
			return nil, err
		}
	}

	opts := chartutil.ReleaseOptions{}
	values, err := chartutil.ToRenderValues(chart, vals, opts, nil)
	if err != nil {
		return nil, err
	}

	templates, err := engine.Render(chart, values)
	if err != nil {
		return nil, err
	}

	var files []*types.File
	for path, template := range templates {
		var file *types.File
		path = normalizePath(h.Path, path, chart.Name())

		file, err = types.NewFileFromPath(path)
		if err != nil {
			file = types.NewFile(path)
		}

		nodes, err := yaml.FromString(template)
		if err != nil {
			return nil, err
		}

		for i := range nodes {
			obj := types.NewObject(nodes[i])
			if file.ResourceExists(obj.ID()) {
				file.Resources[obj.ID()].Rendered = obj
			} else {
				file.Resources[obj.ID()] = &types.Resource{
					Rendered: obj,
					Raw:      obj,
				}
			}
		}

		files = append(files, file)
	}

	return files, nil
}

func (h *Helm) IsValidPath() bool {
	_, err := loader.Load(h.Path)
	if err != nil {
		return false
	}
	return true
}

func normalizePath(basePath, path, chartName string) string {
	return filepath.Join(basePath, strings.TrimPrefix(path, chartName))
}
