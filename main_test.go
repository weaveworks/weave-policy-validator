package main

import (
	"context"
	"testing"

	"github.com/MagalixTechnologies/policy-core/validation"
	"github.com/MagalixTechnologies/weave-iac-validator/internal/policy"
	"github.com/MagalixTechnologies/weave-iac-validator/internal/validator"
	"github.com/stretchr/testify/assert"
)

func TestScanMultipleDirs(t *testing.T) {
	tests := []struct {
		path           string
		valuesFile     string
		policiesPath   string
		resourceCount  int
		violationCount int
	}{
		{
			path:           "tests/data/entities",
			valuesFile:     "values-dev.yaml",
			policiesPath:   "tests/data/policies/kubernetes",
			resourceCount:  10,
			violationCount: 27,
		},
	}

	for _, test := range tests {
		ctx := context.Background()
		files, err := scan(ctx, KustomizationConf{
			Path:           test.path,
			HelmValuesFile: test.valuesFile,
		})

		if err != nil {
			t.Fatalf("unexpected error, %v", err)
		}

		policyKustomizer, err := getKustomizer(KustomizationConf{Path: test.policiesPath})
		if err != nil {
			t.Fatalf("unexpected error, %v", err)
		}

		policySource := policy.NewFilesystemSource(policyKustomizer)
		opaValidator := validation.NewOPAValidator(policySource, false, trigger)
		validator := validator.NewValidator(opaValidator, false)

		result, err := validator.Validate(ctx, files)
		if err != nil {
			t.Fatalf("unexpected error, %v", err)
		}

		assert.Equal(t, test.resourceCount, result.Scanned)
		assert.Equal(t, test.violationCount, result.ViolationCount)
	}
}
