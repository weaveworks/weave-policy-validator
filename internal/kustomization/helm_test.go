package kustomization

import (
	"context"
	"testing"

	"github.com/MagalixTechnologies/policy-core/domain"
	"github.com/stretchr/testify/assert"
)

func TestHelmKustomizer(t *testing.T) {
	tests := []struct {
		path          string
		valuesFile    string
		fileCount     int
		resourceCount int
		entities      map[string]domain.Entity
		policies      map[string]domain.Policy
	}{
		{
			path:       "../../tests/data/entities/helm",
			valuesFile: "../../tests/data/entities/helm/values-dev.yaml",
			fileCount:  1,
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
			path:       "../../tests/data/entities/helm",
			valuesFile: "../../tests/data/entities/helm/values-prod.yaml",
			fileCount:  1,
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
			path:       "../../tests/data/policies/helm",
			valuesFile: "../../tests/data/entities/helm/values-dev.yaml",
			fileCount:  3,
			policies: map[string]domain.Policy{
				"magalix.com/v1/Policy/[noNamespace]/magalix.policies.containers-minimum-replica-count": {
					ID:   "magalix.policies.containers-minimum-replica-count",
					Name: "Containers Minimum Replica Count",
					Parameters: []domain.PolicyParameters{
						{Value: 1},
					},
				},
				"magalix.com/v1/Policy/[noNamespace]/magalix.policies.containers-running-with-privilege-escalation": {
					ID:   "magalix.policies.containers-running-with-privilege-escalation",
					Name: "Containers Running With Privilege Escalation",
					Parameters: []domain.PolicyParameters{
						{Value: true},
					},
				},
				"magalix.com/v1/Policy/[noNamespace]/magalix.policies.containers-running-in-privileged-mode": {
					ID:   "magalix.policies.containers-running-in-privileged-mode",
					Name: "Containers Running In Privileged Mode",
					Parameters: []domain.PolicyParameters{
						{Value: true},
					},
				},
			},
		},
		{
			path:       "../../tests/data/policies/helm",
			valuesFile: "../../tests/data/entities/helm/values-prod.yaml",
			fileCount:  3,
			policies: map[string]domain.Policy{
				"magalix.com/v1/Policy/[noNamespace]/magalix.policies.containers-minimum-replica-count": {
					ID:   "magalix.policies.containers-minimum-replica-count",
					Name: "Containers Minimum Replica Count",
					Parameters: []domain.PolicyParameters{
						{Value: 2},
					},
				},
				"magalix.com/v1/Policy/[noNamespace]/magalix.policies.containers-running-with-privilege-escalation": {
					ID:   "magalix.policies.containers-running-with-privilege-escalation",
					Name: "Containers Running With Privilege Escalation",
					Parameters: []domain.PolicyParameters{
						{Value: false},
					},
				},
				"magalix.com/v1/Policy/[noNamespace]/magalix.policies.containers-running-in-privileged-mode": {
					ID:   "magalix.policies.containers-running-in-privileged-mode",
					Name: "Containers Running In Privileged Mode",
					Parameters: []domain.PolicyParameters{
						{Value: false},
					},
				},
			},
		},
	}

	for _, test := range tests {
		kustomizer := NewHelmKustomizer(test.path)
		kustomizer.SetValueFile(test.valuesFile)

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
					assert.Equal(t, policy.Parameters[0].Value, testPolicy.Parameters[0].Value)

				}
			}
		}
	}
}