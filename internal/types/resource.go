package types

type RemediationHint struct {
	ViolatingKey     *string
	RecommendedValue interface{}
}
type Resource struct {
	Remediated bool
	Rendered   *Object
	Raw        *Object
}

func (r *Resource) FindKey(key string) (int, int) {
	obj := r.Raw
	if obj == nil {
		obj = r.Rendered
	}
	if f := obj.GetField(key); f != nil {
		return f.Key.StartLine, f.Value.EndLine
	}
	return obj.node.StartLine, obj.node.EndLine
}

func (r *Resource) Remediate(key string, value interface{}) (bool, error) {
	if r.Raw == nil {
		return false, nil
	}

	err := r.Raw.SetField(key, value)
	if err != nil {
		return false, err
	}

	return true, nil
}
