package ast

// NodeType represents the type of AST node.
type NodeType string

const (
	NodeTypeStrategy   NodeType = "strategy"
	NodeTypeIndicator  NodeType = "indicator"
	NodeTypeExpression NodeType = "expression"
	NodeTypeCondition  NodeType = "condition"
	NodeTypeEntry      NodeType = "entry"
	NodeTypeExit       NodeType = "exit"
	NodeTypeRisk       NodeType = "risk"
)

// Node represents an abstract syntax tree node.
type Node struct {
	Type       NodeType
	Name       string
	Children   []*Node
	Properties map[string]interface{}
	Parent     *Node
}

// NewNode creates a new AST node.
func NewNode(nodeType NodeType, name string) *Node {
	return &Node{
		Type:       nodeType,
		Name:       name,
		Children:   make([]*Node, 0),
		Properties: make(map[string]interface{}),
	}
}

// AddChild adds a child node.
func (n *Node) AddChild(child *Node) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

// SetProperty sets a node property.
func (n *Node) SetProperty(key string, value interface{}) {
	n.Properties[key] = value
}

// GetProperty gets a node property.
func (n *Node) GetProperty(key string) interface{} {
	return n.Properties[key]
}

// StrategyAST represents the complete strategy AST.
type StrategyAST struct {
	Root *Node
}

// NewStrategyAST creates a new strategy AST.
func NewStrategyAST(root *Node) *StrategyAST {
	return &StrategyAST{Root: root}
}
