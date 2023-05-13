package source

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/weaveworks/weave-policy-validator/internal/types"
	"github.com/weaveworks/weave-policy-validator/internal/yaml"

	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resource"
	ktypes "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type KustomizationFile struct {
	Node             *yaml.Node
	Object           ktypes.Kustomization
	OriginAnnotation bool
}

type Kustomize struct {
	Path   string
	source *krusty.Kustomizer
	fs     filesys.FileSystem
	k      ktypes.Kustomization
}

func NewKustomizeSource(path string) *Kustomize {
	opts := krusty.MakeDefaultOptions()
	return &Kustomize{
		Path:   path,
		source: krusty.MakeKustomizer(opts),
		fs:     filesys.MakeFsOnDisk(),
	}
}

func (k *Kustomize) Type() string {
	return KustomizeType
}

func (k *Kustomize) ResourceFiles(_ context.Context) ([]*types.File, error) {
	kustomizeFile, err := parseKustomizationFile(k.Path)
	if err != nil {
		return nil, err
	}

	resmap, err := k.source.Run(k.fs, k.Path)
	if err != nil {
		return nil, err
	}

	filesMap := make(map[string]*types.File)

	for _, patch := range kustomizeFile.Object.PatchesStrategicMerge {
		path := filepath.Join(k.Path, string(patch))
		file, err := types.NewFileFromPath(path)
		if err != nil {
			return nil, err
		}
		filesMap[path] = file
	}

	resources := make([]*resource.Resource, 0)
	for _, resource := range resmap.Resources() {
		origin, err := resource.GetOrigin()
		if err != nil {
			return nil, err
		}
		resources = append(resources, resource)
		path := filepath.Join(k.Path, origin.Path)

		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		if info.IsDir() {
			continue
		}

		file, err := types.NewFileFromPath(path)
		if err != nil {
			return nil, err
		}
		filesMap[path] = file
	}

	if !kustomizeFile.OriginAnnotation {
		resmap.RemoveOriginAnnotations()
	}

	var files []*types.File
	for _, file := range filesMap {
		files = append(files, file)
	}

	for _, resource := range resources {
		nodes, err := yaml.StringParse(resource.String())
		if err != nil {
			return nil, err
		}
		for i := range nodes {
			obj := types.NewObject(nodes[i])
			for _, file := range files {
				if file.ResourceExists(k.getObjOriginalID(obj.ID())) {
					file.SetResourceRenderedObject(obj)
					break
				}
			}
		}
	}
	return files, nil
}

func (k *Kustomize) getObjOriginalID(id string) string {
	if k.k.NamePrefix != "" {
		id = strings.TrimPrefix(id, k.k.NamePrefix)
	}
	if k.k.NameSuffix != "" {
		id = strings.TrimSuffix(id, k.k.NameSuffix)
	}
	return id
}

func (k *Kustomize) IsValidPath() bool {
	info, err := os.Stat(k.Path)
	if err != nil {
		return false
	}

	if info.IsDir() {
		for _, filename := range konfig.RecognizedKustomizationFileNames() {
			filePath := filepath.Join(k.Path, filename)
			if _, err := os.Stat(filePath); err == nil {
				return true
			}
		}
		return false
	}

	return false
}

func parseKustomizationFile(path string) (*KustomizationFile, error) {
	path = filepath.Join(path, konfig.DefaultKustomizationFileName())

	in, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	nodes, err := yaml.BytesParse(in)
	if err != nil {
		return nil, err
	}

	node := nodes[0]

	var obj ktypes.Kustomization
	err = yaml.Unmarshal(in, &obj)
	if err != nil {
		return nil, err
	}

	var originAnnotations bool
	for i := range obj.BuildMetadata {
		if obj.BuildMetadata[i] == ktypes.OriginAnnotations {
			originAnnotations = true
			break
		}
	}

	if !originAnnotations {
		obj.BuildMetadata = append(obj.BuildMetadata, ktypes.OriginAnnotations)

		out, err := yaml.Marshal(obj)
		if err != nil {
			return nil, err
		}

		err = os.WriteFile(path, out, 0644)
		if err != nil {
			return nil, err
		}
	}

	return &KustomizationFile{
		Node:             node,
		Object:           obj,
		OriginAnnotation: originAnnotations,
	}, nil
}
