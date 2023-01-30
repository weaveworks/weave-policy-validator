package kustomization

import (
	"context"
	"os"
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

func (h *Helm) ResourceFiles(ctx context.Context) ([]*types.File, error) {
	paths, err := glob(h.Path)
	if err != nil {
		return nil, err
	}

	var files []*types.File
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		if !info.IsDir() {
			continue
		}

		if chartFiles, err := h.resourceFiles(ctx, path); err == nil {
			files = append(files, chartFiles...)
		}
	}

	return files, nil
}

func (h *Helm) resourceFiles(_ context.Context, chartPath string) ([]*types.File, error) {
	chart, err := loader.Load(chartPath)
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
		path = normalizePath(chartPath, path, chart.Name())

		nodes, err := yaml.StringParse(template)
		if err != nil {
			return nil, err
		}

		file := types.NewFile(path)
		for i := range nodes {
			obj := types.NewObject(nodes[i])
			file.Resources[obj.ID()] = &types.Resource{
				Rendered: obj,
			}
		}
		files = append(files, file)
	}
	return files, nil
}

func (h *Helm) IsValidPath() bool {
	info, err := os.Stat(h.Path)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return false
	}
	paths, err := glob(h.Path)
	if err != nil {
		return false
	}
	for _, path := range paths {
		if _, err := loader.Load(path); err == nil {
			return true
		}
	}
	return false
}

func normalizePath(basePath, path, chartName string) string {
	return filepath.Join(basePath, strings.TrimPrefix(path, chartName))
}
