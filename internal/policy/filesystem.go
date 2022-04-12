package policy

import (
	"context"
	"fmt"

	"github.com/MagalixTechnologies/policy-core/domain"
	"github.com/MagalixTechnologies/weave-iac-validator/internal/kustomization"
)

type FilesystemPolicySource struct {
	kustomizer kustomization.Kustomizer
}

// NewFilesystemSource creates new Policy filesystem source
func NewFilesystemSource(kustomizer kustomization.Kustomizer) *FilesystemPolicySource {
	return &FilesystemPolicySource{
		kustomizer: kustomizer,
	}
}

// GetAll gets all policies
func (l *FilesystemPolicySource) GetAll(ctx context.Context) ([]domain.Policy, error) {
	files, err := l.kustomizer.ResourceFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get kustomization resources, error: %v", err)
	}
	var policies []domain.Policy
	for _, file := range files {
		for _, resource := range file.Resources {
			if resource.Rendered == nil {
				continue
			}
			policy, err := resource.Rendered.Policy()
			if err != nil {
				return nil, err
			}
			policies = append(policies, policy)
		}
	}
	return policies, nil
}
