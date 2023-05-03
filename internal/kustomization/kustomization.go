package kustomization

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

type Kustomizer interface {
	Type() string
	IsValidPath() bool
	ResourceFiles(context.Context) ([]*types.File, error)
}

func GetKustomizerFromPath(path string) (Kustomizer, error) {
	helm := NewHelmKustomizer(path)
	if helm.IsValidPath() {
		return helm, nil
	}

	kustomize := NewKustomizeKustomizer(path)
	if kustomize.IsValidPath() {
		return kustomize, nil
	}

	kubernetes := NewKubernetesKustomizer(path)
	if kubernetes.IsValidPath() {
		return kubernetes, nil
	}

	return nil, errors.New("path is not recognized as a kustomization valid path")
}
