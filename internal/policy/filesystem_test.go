package policy

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/policy-agent/pkg/policy-core/domain"
	"github.com/weaveworks/weave-policy-validator/internal/kustomization"
)

func TestFileSystemPolicySource(t *testing.T) {
	tests := []struct {
		name           string
		kustomizer     kustomization.Kustomizer
		helmValuesFile string
		policies       map[string]domain.Policy
	}{
		{
			name:       "kubernetes",
			kustomizer: kustomization.NewKubernetesKustomizer("../../tests/data/policies/kubernetes"),
			policies: map[string]domain.Policy{
				"magalix.policies.containers-minimum-replica-count": {
					ID: "magalix.policies.containers-minimum-replica-count",
					Parameters: []domain.PolicyParameters{
						{
							Name:  "replica_count",
							Value: 2,
						},
					},
				},
				"magalix.policies.containers-running-with-privilege-escalation": {
					ID: "magalix.policies.containers-running-with-privilege-escalation",
					Parameters: []domain.PolicyParameters{
						{
							Name:  "allow_privilege_escalation",
							Value: false,
						},
					},
				},
				"magalix.policies.containers-running-in-privileged-mode": {
					ID: "magalix.policies.containers-running-in-privileged-mode",
					Parameters: []domain.PolicyParameters{
						{
							Name:  "privilege",
							Value: false,
						},
					},
				},
			},
		},
		{
			name:           "helm with values-dev values file",
			kustomizer:     kustomization.NewHelmKustomizer("../../tests/data/policies/helm"),
			helmValuesFile: "../../tests/data/policies/helm/values-dev.yaml",
			policies: map[string]domain.Policy{
				"magalix.policies.containers-minimum-replica-count": {
					ID: "magalix.policies.containers-minimum-replica-count",
					Parameters: []domain.PolicyParameters{
						{
							Name:  "replica_count",
							Value: 1,
						},
					},
				},
				"magalix.policies.containers-running-with-privilege-escalation": {
					ID: "magalix.policies.containers-running-with-privilege-escalation",
					Parameters: []domain.PolicyParameters{
						{
							Name:  "allow_privilege_escalation",
							Value: true,
						},
					},
				},
				"magalix.policies.containers-running-in-privileged-mode": {
					ID: "magalix.policies.containers-running-in-privileged-mode",
					Parameters: []domain.PolicyParameters{
						{
							Name:  "privilege",
							Value: true,
						},
					},
				},
			},
		},
		{
			name:           "helm with values-prod values file",
			kustomizer:     kustomization.NewHelmKustomizer("../../tests/data/policies/helm"),
			helmValuesFile: "../../tests/data/policies/helm/values-prod.yaml",
			policies: map[string]domain.Policy{
				"magalix.policies.containers-minimum-replica-count": {
					ID: "magalix.policies.containers-minimum-replica-count",
					Parameters: []domain.PolicyParameters{
						{
							Name:  "replica_count",
							Value: 2,
						},
					},
				},
				"magalix.policies.containers-running-with-privilege-escalation": {
					ID: "magalix.policies.containers-running-with-privilege-escalation",
					Parameters: []domain.PolicyParameters{
						{
							Name:  "allow_privilege_escalation",
							Value: false,
						},
					},
				},
				"magalix.policies.containers-running-in-privileged-mode": {
					ID: "magalix.policies.containers-running-in-privileged-mode",
					Parameters: []domain.PolicyParameters{
						{
							Name:  "privilege",
							Value: false,
						},
					},
				},
			},
		},
		{
			name:       "kustomize with dev overlay",
			kustomizer: kustomization.NewKustomizeKustomizer("../../tests/data/policies/kustomize/overlays/dev"),
			policies: map[string]domain.Policy{
				"magalix.policies.containers-minimum-replica-count": {
					ID: "magalix.policies.containers-minimum-replica-count",
					Parameters: []domain.PolicyParameters{
						{
							Name:  "replica_count",
							Value: 1,
						},
					},
				},
				"magalix.policies.containers-running-with-privilege-escalation": {
					ID: "magalix.policies.containers-running-with-privilege-escalation",
					Parameters: []domain.PolicyParameters{
						{
							Name:  "allow_privilege_escalation",
							Value: true,
						},
					},
				},
				"magalix.policies.containers-running-in-privileged-mode": {
					ID: "magalix.policies.containers-running-in-privileged-mode",
					Parameters: []domain.PolicyParameters{
						{
							Name:  "privilege",
							Value: true,
						},
					},
				},
			},
		},
		{
			name:       "kustomize with prod overlay",
			kustomizer: kustomization.NewKustomizeKustomizer("../../tests/data/policies/kustomize/overlays/prod"),
			policies: map[string]domain.Policy{
				"magalix.policies.containers-minimum-replica-count": {
					ID: "magalix.policies.containers-minimum-replica-count",
					Parameters: []domain.PolicyParameters{
						{
							Name:  "replica_count",
							Value: 2,
						},
					},
				},
				"magalix.policies.containers-running-with-privilege-escalation": {
					ID: "magalix.policies.containers-running-with-privilege-escalation",
					Parameters: []domain.PolicyParameters{
						{
							Name:  "allow_privilege_escalation",
							Value: false,
						},
					},
				},
				"magalix.policies.containers-running-in-privileged-mode": {
					ID: "magalix.policies.containers-running-in-privileged-mode",
					Parameters: []domain.PolicyParameters{
						{
							Name:  "privilege",
							Value: false,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		if test.helmValuesFile != "" {
			test.kustomizer.(*kustomization.Helm).SetValueFile(test.helmValuesFile)
		}

		source := NewFilesystemSource(test.kustomizer)
		policies, err := source.GetAll(context.Background())
		if err != nil {
			t.Errorf("failed to get policies, error: %v", err)
		}

		assert.Equal(t, len(policies), len(test.policies))
		for _, policy := range policies {
			assert.Equal(t, test.policies[policy.ID].ID, policy.ID)
			assert.Equal(t, test.policies[policy.ID].Parameters[0].Value, policy.Parameters[0].Value)
		}
	}
}
