package yaml

import (
	"bytes"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var regex = regexp.MustCompile("^([a-zA-Z0-9]+)\\[([0-9]+)\\]")

type Field struct {
	Key   *Node
	Value *Node
}

type Node struct {
	node      *yaml.Node
	StartLine int
	EndLine   int
}

// Marshal serializes the value provided into a YAML document
func Marshal(in interface{}) ([]byte, error) {
	return yaml.Marshal(in)
}

// Unmarshal decodes YAML into the provided object
func Unmarshal(in []byte, out interface{}) error {
	return yaml.Unmarshal(in, out)
}

// SingleDocFromFile load file from path and return its first document
func SingleDocFromFile(path string) (*Node, error) {
	in, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var node yaml.Node
	err = yaml.Unmarshal(in, &node)
	if err != nil {
		return nil, err
	}

	return newNode(&node), nil
}

// MultiDocFromFile load file from path and return all its documents
func MultiDocFromFile(path string) ([]*Node, error) {
	in, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return FromBytes(in)
}

func FromString(in string) ([]*Node, error) {
	return FromBytes([]byte(in))
}

func FromBytes(in []byte) ([]*Node, error) {
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
		nodes = append(nodes, newNode(&node))
	}
	return nodes, nil
}

func Bytes(nodes []*Node) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	for _, node := range nodes {
		err := encoder.Encode(node.node)
		if err != nil {
			return nil, err
		}
	}

	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

func (n *Node) Value() string {
	return n.node.Value
}

func (n *Node) GetField(path string, exact bool) *Field {
	knode, vnode := getByKeyPath(n.node, path, false, exact)
	if knode == nil || vnode == nil {
		return nil
	}
	return &Field{
		Key:   newNode(knode),
		Value: newNode(vnode),
	}
}

func (n *Node) SetField(path string, value interface{}) error {
	_, vnode := getByKeyPath(n.node, path, true, false)
	return vnode.Encode(value)
}

func (n *Node) Map() (map[string]interface{}, error) {
	var m map[string]interface{}
	err := n.node.Decode(&m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func newNode(node *yaml.Node) *Node {
	return &Node{
		node:      node,
		StartLine: node.Line,
		EndLine:   lastChild(node).Line,
	}
}

func lastChild(n *yaml.Node) *yaml.Node {
	if len(n.Content) == 0 {
		return n
	}
	n = n.Content[len(n.Content)-1]
	return lastChild(n)
}

func getByKey(n *yaml.Node, key string, create bool) (*yaml.Node, *yaml.Node) {
	if n.Kind == yaml.MappingNode {
		for i := 0; i < len(n.Content); i += 2 {
			if key == n.Content[i].Value {
				return n.Content[i], n.Content[i+1]
			}
		}
		if create {
			line := lastChild(n).Line + 1
			keyNode := &yaml.Node{
				Value: key,
				Kind:  yaml.ScalarNode,
				Line:  line,
			}
			valueNode := &yaml.Node{
				Kind: yaml.MappingNode,
				Line: line,
			}
			n.Content = append(n.Content, keyNode, valueNode)
			return keyNode, valueNode
		}
	}
	for i := range n.Content {
		return getByKey(n.Content[i], key, create)
	}
	return nil, nil
}

func getByKeyPath(node *yaml.Node, path string, create bool, exact bool) (*yaml.Node, *yaml.Node) {
	rootKeyNode := &yaml.Node{}
	rootValueNode := node

	keys := parseKeyPath(path)
	for _, key := range keys {
		index, err := strconv.ParseInt(key, 0, 10)
		if err == nil {
			if rootValueNode.Kind == yaml.SequenceNode && int(index) < len(rootValueNode.Content) {
				rootKeyNode = rootValueNode.Content[index]
				rootValueNode = rootValueNode.Content[index]
			} else {
				if exact {
					return nil, nil
				}
				break
			}
		} else {
			keyNode, valueNode := getByKey(rootValueNode, key, create)
			if keyNode == nil || valueNode == nil {
				if exact {
					return nil, nil
				}
				break
			}
			rootKeyNode = keyNode
			rootValueNode = valueNode
		}
	}
	return rootKeyNode, rootValueNode
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
