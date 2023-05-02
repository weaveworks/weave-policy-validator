package kustomization

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/policy-agent/pkg/policy-core/domain"
)

func TestKubernetesKustomizer(t *testing.T) {
	tests := []struct {
		path          string
		fileCount     int
		resourceCount int
		entities      map[string]domain.Entity
		policies      map[string]domain.Policy
	}{
		{
			path:      "../../tests/data/entities/kubernetes",
			fileCount: 1,
			entities: map[string]domain.Entity{
				"apps/v1/Deployment/[noNamespace]/frontend": {
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "frontend",
				},
				"apps/v1/Deployment/[noNamespace]/backend": {
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "backend",
				},
			},
		},
		{
			path:      "../../tests/data/policies/kubernetes",
			fileCount: 3,
			policies: map[string]domain.Policy{
				"magalix.com/v1/Policy/[noNamespace]/magalix.policies.containers-minimum-replica-count": {
					ID:   "magalix.policies.containers-minimum-replica-count",
					Name: "Containers Minimum Replica Count",
				},
				"magalix.com/v1/Policy/[noNamespace]/magalix.policies.containers-running-with-privilege-escalation": {
					ID:   "magalix.policies.containers-running-with-privilege-escalation",
					Name: "Containers Running With Privilege Escalation",
				},
				"magalix.com/v1/Policy/[noNamespace]/magalix.policies.containers-running-in-privileged-mode": {
					ID:   "magalix.policies.containers-running-in-privileged-mode",
					Name: "Containers Running In Privileged Mode",
				},
			},
		},
	}

	for _, test := range tests {
		kustomizer := NewKubernetesKustomizer(test.path)
		files, err := kustomizer.ResourceFiles(context.Background())
		if err != nil {
			t.Errorf("failed to get resouces, error: %v", err)
		}

		assert.Equal(t, len(files), test.fileCount)

		for _, file := range files {
			for _, resource := range file.Resources {
				if len(test.entities) > 0 {
					entity, err := resource.Rendered.Entity()
					if err != nil {
						t.Errorf("failed to get entity, error: %v", err)
					}
					testEntity := test.entities[resource.Rendered.ID()]
					assert.Equal(t, entity.APIVersion, testEntity.APIVersion)
					assert.Equal(t, entity.Kind, testEntity.Kind)
					assert.Equal(t, entity.Namespace, testEntity.Namespace)
					assert.Equal(t, entity.Name, testEntity.Name)
				}

				if len(test.policies) > 0 {
					policy, err := resource.Rendered.Policy()
					if err != nil {
						t.Errorf("failed to get policy, error: %v", err)
					}
					testPolicy := test.policies[resource.Rendered.ID()]
					assert.Equal(t, policy.ID, testPolicy.ID)
					assert.Equal(t, policy.Name, testPolicy.Name)
				}
			}
		}
	}
}
