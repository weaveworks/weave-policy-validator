package policy

import (
	"context"
	"fmt"

	"github.com/weaveworks/policy-agent/pkg/policy-core/domain"
	"github.com/weaveworks/weave-policy-validator/internal/source"
)

type FilesystemPolicySource struct {
	source source.Source
}

// NewFilesystemSource creates new Policy filesystem source
func NewFilesystemSource(source source.Source) *FilesystemPolicySource {
	return &FilesystemPolicySource{
		source: source,
	}
}

// GetAll gets all policies
func (l *FilesystemPolicySource) GetAll(ctx context.Context) ([]domain.Policy, error) {
	files, err := l.source.ResourceFiles(ctx)
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

func (l *FilesystemPolicySource) GetPolicyConfig(ctx context.Context, entity domain.Entity) (*domain.PolicyConfig, error) {
	return nil, nil
}
