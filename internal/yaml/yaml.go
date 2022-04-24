package yaml

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var regex = regexp.MustCompile("^([a-zA-Z0-9]+)\\[([0-9]+)\\]")

type Node struct {
	*yaml.RNode
}

func newNode(rn *yaml.RNode) *Node {
	return &Node{rn}
}

// GetField gets field's node by json path
func (n *Node) GetField(path string) (*Node, error) {
	fields := parseKeyPath(path)
	pathGetter := yaml.Lookup(fields...)

	rn, err := n.Pipe(pathGetter)
	if err != nil {
		return nil, err
	}

	if rn == nil {
		return nil, nil
	}

	return newNode(rn), nil
}

// FindField finds field's node or its nearest parent by json path
func (n *Node) FindField(path string) (*Node, error) {
	fields := parseKeyPath(path)
	for i := range fields {
		pathGetter := yaml.Lookup(fields[:len(fields)-i]...)
		rn, err := n.Pipe(pathGetter)
		if err != nil {
			return nil, err
		}
		if rn != nil {
			return newNode(rn), nil
		}
	}
	return nil, nil
}

// SetField sets field value
func (n *Node) SetField(path string, value interface{}) error {
	fields := parseKeyPath(path)
	pathGetter := yaml.LookupCreate(yaml.MappingNode, fields...)

	ncopy := n.Copy()
	node, err := ncopy.Pipe(pathGetter)
	if err != nil {
		return err
	}

	if node == nil {
		return fmt.Errorf("cannot find field: %s", path)
	}

	node, _ = n.Pipe(pathGetter)
	return node.Document().Encode(value)
}

// Marshal serializes the value provided into a YAML document
func Marshal(in interface{}) ([]byte, error) {
	return yaml.Marshal(in)
}

// Unmarshal decodes YAML into the provided object
func Unmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}

// SingleDocFromFile loads file from path and returns its first document
func SingleDocFromFile(path string) (*Node, error) {
	in, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var node yaml.RNode
	err = yaml.Unmarshal(in, &node)
	if err != nil {
		return nil, err
	}

	return newNode(&node), nil
}

// MultiDocFromFile loads file from path and returns all its documents
func MultiDocFromFile(path string) ([]*Node, error) {
	in, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return BytesParse(in)
}

// StringParse loads yaml documents from string
func StringParse(in string) ([]*Node, error) {
	return BytesParse([]byte(in))
}

// BytesParse loads yaml documents from bytes
func BytesParse(in []byte) ([]*Node, error) {
	var nodes []*Node
	reader := bytes.NewReader(in)
	decoder := yaml.NewDecoder(reader)
	for {
		var node yaml.Node
		if err := decoder.Decode(&node); err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		nodes = append(nodes, newNode(yaml.NewRNode(&node)))
	}
	return nodes, nil
}

// Bytes encodes multiple yaml documents into bytes
func Bytes(nodes []*Node) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	for _, node := range nodes {
		err := encoder.Encode(node.Document())
		if err != nil {
			return nil, err
		}
	}

	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

func parseKeyPath(path string) []string {
	var keys []string
	parts := strings.Split(path, ".")
	for _, part := range parts {
		groups := regex.FindStringSubmatch(part)
		if groups == nil {
			keys = append(keys, part)
		} else {
			keys = append(keys, groups[1:]...)
		}
	}
	return keys
}

func lastChild(n *yaml.Node) *yaml.Node {
	if len(n.Content) == 0 {
		return n
	}
	n = n.Content[len(n.Content)-1]
	return lastChild(n)
}

func (n *Node) StartLine() int {
	return n.Document().Line
}

func (n *Node) EndLine() int {
	tail := lastChild(n.Document())
	return tail.Line
}
