package source

import (
	"context"
	"errors"

	"github.com/weaveworks/weave-policy-validator/internal/types"
)

const (
	HelmType       = "helm"
	KustomizeType  = "kustomize"
	KubernetesType = "kubernetes"
)

type Source interface {
	Type() string
	IsValidPath() bool
	ResourceFiles(context.Context) ([]*types.File, error)
}

func GetSourceFromPath(path string) (Source, error) {
	helm := NewHelmSource(path)
	if helm.IsValidPath() {
		return helm, nil
	}

	kustomize := NewKustomizeSource(path)
	if kustomize.IsValidPath() {
		return kustomize, nil
	}

	kubernetes := NewKubernetesSource(path)
	if kubernetes.IsValidPath() {
		return kubernetes, nil
	}

	return nil, errors.New("path is not recognized as a valid path")
}
