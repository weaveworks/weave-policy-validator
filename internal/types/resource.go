package types

import (
	"encoding/json"
)

type RemediationHint struct {
	ViolatingKey     *string
	RecommendedValue interface{}
}
type Resource struct {
	Remediated bool
	Rendered   *Object
	Raw        *Object
}

// FindKey returns key start and end lines
func (r *Resource) FindKey(key string) (int, int) {
	startLine := r.Rendered.node.StartLine()
	endLine := r.Rendered.node.EndLine()

	if r.Raw == nil {
		return startLine, endLine
	}

	field, err := r.Raw.node.FindField(key)
	if err != nil || field == nil {
		return startLine, endLine
	}

	return field.StartLine(), field.EndLine()
}

// Remediate rremediates resource value
func (r *Resource) Remediate(key string, value interface{}) (bool, error) {
	if number, ok := value.(json.Number); ok {
		value, _ = number.Float64()
	}

	if r.Raw == nil {
		return false, nil
	}

	err := r.Raw.SetField(key, value)
	if err != nil {
		return false, err
	}

	return true, nil
}
