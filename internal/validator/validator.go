package validator

import (
	"context"

	"github.com/weaveworks/policy-agent/pkg/policy-core/domain"
	"github.com/weaveworks/policy-agent/pkg/policy-core/validation"
	"github.com/weaveworks/weave-iac-validator/internal/types"
)

type Validator struct {
	validator validation.Validator
	remediate bool
}

// NewValidator return new validator struct
func NewValidator(validator validation.Validator, remediate bool) *Validator {
	return &Validator{
		validator: validator,
		remediate: remediate,
	}
}

// Validate validates resources against policies
func (v *Validator) Validate(ctx context.Context, files []*types.File) (*types.Result, error) {
	results := types.Result{
		Violations: []types.Violation{},
	}

	for _, file := range files {
		for _, resource := range file.Resources {
			if resource.Rendered == nil {
				continue
			}

			entity, err := resource.Rendered.Entity()
			if err != nil {
				return nil, err
			}

			summary, err := v.validator.Validate(ctx, entity, "")
			if err != nil {
				return nil, err
			}

			for _, violation := range summary.Violations {
				result := types.Violation{
					ID:      violation.ID,
					Message: violation.Message,
					Policy: types.Policy{
						ID:          violation.Policy.ID,
						Name:        violation.Policy.Name,
						Severity:    violation.Policy.Severity,
						Category:    violation.Policy.Category,
						Description: violation.Policy.Description,
						HowToSolve:  violation.Policy.HowToSolve,
					},
					Entity: types.Entity{
						Name:      entity.Name,
						Namespace: entity.Namespace,
						Kind:      entity.Kind,
					},
					Details: getDetails(violation.Occurrences),
				}

				startLine, endLine := 1, 1
				if result.Details.ViolatingKey != nil {
					startLine, endLine = resource.FindKey(*result.Details.ViolatingKey)
					if endLine < startLine {
						endLine = startLine
					}

					if v.remediate && result.Details.RecommendedValue != nil {
						remediated, err := resource.Remediate(*result.Details.ViolatingKey, result.Details.RecommendedValue)
						if err == nil && remediated {
							file.Remediated = true
							resource.Remediated = true
							results.Remediated++
						}
					}
				}

				result.Location = types.Location{
					Path:      file.Path,
					StartLine: startLine,
					EndLine:   endLine,
				}

				results.Violations = append(results.Violations, result)
				results.ViolationCount++
			}
			results.Scanned++
		}
	}
	return &results, nil
}

// @todo check if we will handle auto remediating multiple occurences
func getDetails(occurrences []domain.Occurrence) types.Details {
	details := types.Details{}
	if len(occurrences) < 1 {
		return details
	}
	details.ViolatingKey = occurrences[0].ViolatingKey
	details.RecommendedValue = occurrences[0].RecommendedValue

	return details
}
