package types

import (
	"strings"

	"github.com/MagalixTechnologies/policy-core/domain"
	"github.com/MagalixTechnologies/weave-iac-validator/internal/yaml"
)

const (
	ApiVersionField = "apiVersion"
	KindField       = "kind"
	NamespaceField  = "metadata.namespace"
	NameField       = "metadata.name"
	SpecField       = "spec"
	seperator       = "/"
)

type Object struct {
	node *yaml.Node
}

// NewObject create new object
func NewObject(node *yaml.Node) *Object {
	return &Object{node: node}
}

// ApiVersion gets apiVersion
func (obj *Object) ApiVersion() string {
	return obj.getFieldValue(ApiVersionField)
}

// Kind gets object kind
func (obj *Object) Kind() string {
	return obj.getFieldValue(KindField)
}

// Namespace gets object namespace
func (obj *Object) Namespace() string {
	namespace := obj.getFieldValue(NamespaceField)
	if namespace == "" {
		return "[noNamespace]"
	}
	return namespace
}

// Name gets object name
func (obj *Object) Name() string {
	return obj.getFieldValue(NameField)
}

// ID get object id
func (obj *Object) ID() string {
	parts := []string{
		obj.ApiVersion(),
		obj.Kind(),
		obj.Namespace(),
		obj.Name(),
	}
	return strings.Join(parts, seperator)
}

// GetField get field from key path
func (obj *Object) GetField(key string) *yaml.Field {
	return obj.node.GetField(key, true)
}

// GetNearestField field or its nearest parent from key path
func (obj *Object) GetNearestField(key string) *yaml.Field {
	return obj.node.GetField(key, false)
}

// SetField set field value
func (obj *Object) SetField(key string, value interface{}) error {
	return obj.node.SetField(key, value)
}

// Entity convert object to entity
func (obj *Object) Entity() (domain.Entity, error) {
	spec, err := obj.node.Map()
	if err != nil {
		return domain.Entity{}, nil
	}
	return domain.NewEntityFromSpec(spec), nil
}

// Policy convert object to policy
func (obj *Object) Policy() (domain.Policy, error) {
	var policy domain.Policy

	m, err := obj.node.Map()
	if err != nil {
		return policy, nil
	}

	spec := m[SpecField]
	in, err := yaml.Marshal(spec)
	if err != nil {
		return policy, nil
	}

	err = yaml.Unmarshal(in, &policy)
	return policy, err
}

func (obj *Object) getFieldValue(key string) string {
	var value string
	if f := obj.node.GetField(key, true); f != nil {
		value = f.Value.Value()
	}
	return value
}
