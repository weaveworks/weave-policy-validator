package validator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/policy-agent/pkg/policy-core/validation"
	"github.com/weaveworks/weave-policy-validator/internal/policy"
	"github.com/weaveworks/weave-policy-validator/internal/source"
	"github.com/weaveworks/weave-policy-validator/internal/types"
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
		entitySource, err := source.GetSourceFromPath(test.path)
		if err != nil {
			t.Error(err)
		}
		if entitySource.Type() == source.HelmType {
			entitySource.(*source.Helm).SetValueFile(test.helmValuesPath)
		}

		policySource, err := source.GetSourceFromPath(test.policiesPath)
		if err != nil {
			t.Error(err)
		}
		if policySource.Type() == source.HelmType {
			policySource.(*source.Helm).SetValueFile(test.policiesHelmValuesPath)
		}

		fsPolicySource := policy.NewFilesystemSource(policySource)
		opaValidator := validation.NewOPAValidator(fsPolicySource, false, "", "", "", false)
		validator := NewValidator(opaValidator, true)

		ctx := context.Background()
		files, err := entitySource.ResourceFiles(ctx)
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
