package trie

import (
	"strings"
)

type Node struct {
	children map[string]*Node
	pathEnd  bool
}

func newNode() *Node {
	return &Node{
		children: make(map[string]*Node, 0),
	}
}

func NewTrie() *Trie {
	return &Trie{root: newNode()}
}

type Trie struct {
	root *Node
}

func (t *Trie) Insert(path string) {
	current := t.root
	for _, part := range strings.Split(path, "/") {
		if _, ok := current.children[part]; !ok {
			current.children[part] = newNode()
		}
		current = current.children[part]
	}
	current.pathEnd = true
}

func (t *Trie) Search(path string) bool {
	current := t.root
	for _, part := range strings.Split(path, "/") {
		if _, ok := current.children[part]; !ok {
			return false
		} else {
			current = current.children[part]
		}
	}
	return current.pathEnd
}
