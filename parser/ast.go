package parser

type Node interface {
	Children() []Node
	Parent() Node
	append(Node) Node
}

type node struct {
	parent   Node
	children []Node
}

func (n *node) Children() []Node {
	return n.children
}

func (n *node) Parent() Node {
	return n.parent
}

func (n *node) append(c Node) Node {
	n.children = append(n.children, c)
	return c
}

type ListNode struct {
	node
	Not   bool
	Chars string
}

type RangeNode struct {
	node
	Not    bool
	Lo, Hi rune
}

type TextNode struct {
	node
	Text string
}

type PatternNode struct{ node }
type AnyNode struct{ node }
type SuperNode struct{ node }
type SingleNode struct{ node }
type AnyOfNode struct{ node }
