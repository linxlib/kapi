package toml

import "strings"

type Node struct {
	name     string
	value    Value
	kind     Kind
	Children map[string]*Node
	parent   *Node
}

func newNodePointer() *Node {
	output := new(Node)
	output.Children = nil
	output.parent = nil
	return output
}

func (n *Node) createChildren() {
	if n.Children != nil {
		return
	}
	n.Children = make(map[string]*Node)
}

func (n *Node) child(name string) (*Node, bool) {
	if !n.hasChildren() {
		return nil, false
	}
	node, ok := n.Children[name]
	return node, ok
}

func (n *Node) setChild(name string, node *Node) {
	n.createChildren()
	n.Children[name] = node
	node.parent = n
}

func (n *Node) hasChildren() bool {
	return n.Children != nil
}

func (n *Node) String() string {
	output := ""

	if n.kind == kindRoot && n.hasChildren() {
		for _, node := range n.Children {
			output += node.String()
			output += "\n"
		}
	}

	if n.kind == kindSection {
		output += "[" + n.FullName() + "]"
		output += "\n"
		if n.hasChildren() {
			for _, node := range n.Children {
				output += node.String()
			}
		}
	}

	if n.kind == kindValue {
		output += n.name + " = " + n.value.String()
		output += "\n"
	}

	return output
}

func (n *Node) FullName() string {
	output := n.name
	current := n
	for {
		current = current.parent
		if current == nil || current.kind == kindRoot {
			break
		}
		output = current.name + "." + output
	}
	return output
}

func (n *Node) loadValues() {
	n.value, _, _ = parseValue(n.value.raw)

	for _, node := range n.Children {
		node.loadValues()
	}
}

func (n *Node) GetSection(path string) (*Node, bool) {
	names := strings.Split(path, ".")
	current := n
	nameIndex := 0

	for _, node := range current.Children {
		if node.kind != kindSection {
			continue
		}
		if node.name == names[nameIndex] {
			current = node
			nameIndex++
			if nameIndex >= len(names) {
				return current, true
			}
			break
		}
	}

	return current, false
}

func (n *Node) GetValue(path string) (Value, bool) {
	var output Value
	names := strings.Split(path, ".")
	if len(names) == 1 {
		node, ok := n.child(path)
		if !ok {
			return output, false
		}
		return node.value, true
	}

	sectionPath := strings.Join(names[0:len(names)-1], ".")
	section, ok := n.GetSection(sectionPath)
	if !ok {
		return output, false
	}
	return section.GetValue(names[len(names)-1])
}
