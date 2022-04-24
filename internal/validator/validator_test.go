package validator

import (
	"context"
	"testing"

	"github.com/MagalixTechnologies/policy-core/validation"
	"github.com/MagalixTechnologies/weave-iac-validator/internal/kustomization"
	"github.com/MagalixTechnologies/weave-iac-validator/internal/policy"
	"github.com/MagalixTechnologies/weave-iac-validator/internal/types"
	"github.com/stretchr/testify/assert"
)

func TestValidator(t *testing.T) {
	tests := []struct {
		path                   string
		helmValuesPath         string
		policiesPath           string
		policiesHelmValuesPath string
		result                 types.Result
	}{
		{
			path:         "../../tests/data/entities/kubernetes",
			policiesPath: "../../tests/data/policies/kubernetes",
			result: types.Result{
				Scanned:        2,
				ViolationCount: 6,
				Remediated:     6,
			},
		},
		{
			path:                   "../../tests/data/entities/helm",
			helmValuesPath:         "../../tests/data/entities/helm/values-dev.yaml",
			policiesPath:           "../../tests/data/policies/helm",
			policiesHelmValuesPath: "../../tests/data/policies/helm/values-dev.yaml",
			result: types.Result{
				Scanned:        2,
				ViolationCount: 0,
				Remediated:     0,
			},
		},
		{
			path:                   "../../tests/data/entities/helm",
			helmValuesPath:         "../../tests/data/entities/helm/values-prod.yaml",
			policiesPath:           "../../tests/data/policies/helm",
			policiesHelmValuesPath: "../../tests/data/policies/helm/values-prod.yaml",
			result: types.Result{
				Scanned:        2,
				ViolationCount: 0,
				Remediated:     0,
			},
		},
		{
			path:                   "../../tests/data/entities/helm",
			helmValuesPath:         "../../tests/data/entities/helm/values-dev.yaml",
			policiesPath:           "../../tests/data/policies/helm",
			policiesHelmValuesPath: "../../tests/data/policies/helm/values-prod.yaml",
			result: types.Result{
				Scanned:        2,
				ViolationCount: 6,
				Remediated:     0,
			},
		},
		{
			path:                   "../../tests/data/entities/helm",
			helmValuesPath:         "../../tests/data/entities/helm/values-prod.yaml",
			policiesPath:           "../../tests/data/policies/helm",
			policiesHelmValuesPath: "../../tests/data/policies/helm/values-dev.yaml",
			result: types.Result{
				Scanned:        2,
				ViolationCount: 4,
				Remediated:     0,
			},
		},
		{
			path:         "../../tests/data/entities/kustomize/overlays/dev",
			policiesPath: "../../tests/data/policies/kustomize/overlays/dev",
			result: types.Result{
				ViolationCount: 0,
				Remediated:     0,
				Scanned:        2,
			},
		},
		{
			path:         "../../tests/data/entities/kustomize/overlays/prod",
			policiesPath: "../../tests/data/policies/kustomize/overlays/prod",
			result: types.Result{
				ViolationCount: 3,
				Remediated:     3,
				Scanned:        2,
			},
		},
		{
			path:         "../../tests/data/entities/kustomize/overlays/dev",
			policiesPath: "../../tests/data/policies/kustomize/overlays/prod",
			result: types.Result{
				ViolationCount: 6,
				Remediated:     6,
				Scanned:        2,
			},
		},
		{
			path:         "../../tests/data/entities/kustomize/overlays/prod",
			policiesPath: "../../tests/data/policies/kustomize/overlays/dev",
			result: types.Result{
				ViolationCount: 2,
				Remediated:     2,
				Scanned:        2,
			},
		},
	}

	for _, test := range tests {
		entityKustomizer, err := kustomization.GetKustomizerFromPath(test.path)
		if err != nil {
			t.Error(err)
		}
		if entityKustomizer.Type() == kustomization.HelmType {
			entityKustomizer.(*kustomization.Helm).SetValueFile(test.helmValuesPath)
		}

		policyKustomizer, err := kustomization.GetKustomizerFromPath(test.policiesPath)
		if err != nil {
			t.Error(err)
		}
		if policyKustomizer.Type() == kustomization.HelmType {
			policyKustomizer.(*kustomization.Helm).SetValueFile(test.policiesHelmValuesPath)
		}

		policySource := policy.NewFilesystemSource(policyKustomizer)
		opaValidator := validation.NewOPAValidator(policySource, false, "")
		validator := NewValidator(opaValidator, true)

		ctx := context.Background()
		files, err := entityKustomizer.ResourceFiles(ctx)
		if err != nil {
			t.Error(err)
		}

		result, err := validator.Validate(ctx, files)
		if err != nil {
			t.Error(err)
		}

		assert.Equal(t, test.result.Scanned, result.Scanned, "wrong scanned")
		assert.Equal(t, test.result.ViolationCount, result.ViolationCount, "wrong violations")
		assert.Equal(t, test.result.Remediated, result.Remediated, "wrong remediated")
	}
}
