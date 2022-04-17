package validator

import (
	"context"

	"github.com/MagalixTechnologies/policy-core/validation"
	"github.com/MagalixTechnologies/weave-iac-validator/internal/types"
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
					Details: getDetails(violation.Details),
				}

				if result.Details.ViolatingKey != nil {
					startLine, endLine := resource.FindKey(*result.Details.ViolatingKey)
					result.Location = types.Location{
						Path:      file.Path,
						StartLine: startLine,
						EndLine:   endLine,
					}

					if v.remediate && result.Details.RecommendedValue != nil {
						remediated, err := resource.Remediate(*result.Details.ViolatingKey, result.Details.RecommendedValue)
						if err != nil {
							return nil, err
						}
						if remediated {
							file.Remediated = true
							resource.Remediated = true
							results.Remediated++
						}
					}
				}
				results.Violations = append(results.Violations, result)
				results.ViolationCount++
			}
			results.Scanned++
		}
	}
	return &results, nil
}

func getDetails(in map[string]interface{}) types.Details {
	details := types.Details{}
	if key, ok := in["violating_key"]; ok {
		if key, ok := key.(string); ok {
			details.ViolatingKey = &key
		}
	}
	if value, ok := in["recommended_value"]; ok {
		details.RecommendedValue = value
	}
	return details
}
