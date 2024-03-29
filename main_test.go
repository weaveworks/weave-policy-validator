package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/policy-agent/pkg/policy-core/validation"
	"github.com/weaveworks/weave-policy-validator/internal/policy"
	"github.com/weaveworks/weave-policy-validator/internal/validator"
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
		files, err := scan(ctx, SourceConf{
			Path:           test.path,
			HelmValuesFile: test.valuesFile,
		})

		if err != nil {
			t.Fatalf("unexpected error, %v", err)
		}

		policySource, err := getSource(SourceConf{Path: test.policiesPath})
		if err != nil {
			t.Fatalf("unexpected error, %v", err)
		}

		fsPolicySource := policy.NewFilesystemSource(policySource)
		opaValidator := validation.NewOPAValidator(fsPolicySource, false, "", "", "", false)
		validator := validator.NewValidator(opaValidator, false)

		result, err := validator.Validate(ctx, files)
		if err != nil {
			t.Fatalf("unexpected error, %v", err)
		}

		assert.Equal(t, test.resourceCount, result.Scanned)
		assert.Equal(t, test.violationCount, result.ViolationCount)
	}
}
