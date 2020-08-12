package htmlquery

import (
	"bytes"
	"errors"
	"regexp"

	"github.com/antchfx/xpath"
	"golang.org/x/net/html"
)

type Node html.Node

func (n *Node) Regexp(exp string) []string {
	return regexp.MustCompile(exp).FindAllString(n.Data, -1)
}

func (n *Node) RegexpEx(exp string) [][]string {
	return regexp.MustCompile(exp).FindAllStringSubmatch(n.Data, -1)
}

func (n *Node) AttributeValue(key string) (string, error) {
	if n.Type == html.ElementNode && n.Parent == nil && key == n.Data {
		return n.InnerText(), nil
	}
	if attr := n.GetAttributeByKey(key); attr != nil {
		return attr.Val, nil
	}
	return "", errors.New("attribute is nil")
}

func (n *Node) GetAttributeByValue(val string) *Attribute {
	for _, attr := range n.Attr {
		if attr.Val == val {
			return (*Attribute)(&attr)
		}
	}
	return nil
}

func (n *Node) GetAttributeByNamespace(namespace string) *Attribute {
	for _, attr := range n.Attr {
		if attr.Namespace == namespace {
			return (*Attribute)(&attr)
		}
	}
	return nil
}

func (n *Node) Attribute(key string) *Attribute {
	return n.GetAttributeByKey(key)
}

func (n *Node) Attributes() []*Attribute {
	var result []*Attribute
	for i := range n.Attr {
		result = append(result, (*Attribute)(&n.Attr[i]))
	}
	return result
}

func (n *Node) GetAttributeByKey(key string) *Attribute {

	for _, attr := range n.Attr {
		if attr.Key == key {
			return (*Attribute)(&attr)
		}
	}
	return nil
}

func (n *Node) Last() *Node {
	return (*Node)(n.LastChild)
}

func (n *Node) First() *Node {
	return (*Node)(n.FirstChild)
}

func (n *Node) Next() *Node {
	return (*Node)(n.NextSibling)
}

func (n *Node) Prev() *Node {
	return (*Node)(n.PrevSibling)
}

// func (n *Node) NodeName() string {
// 	switch n.Type {
// 	case html.CommentNode:
// 		return "#comment"
// 	case html.DocumentNode:
// 		return "#document"
// 	}
// 	return ""
// }

func (n *Node) TagName() (string, error) {
	if n.Type == html.ElementNode {
		return n.Data, nil
	}
	return "", errors.New("the node is not ElementNode")
}

func (n *Node) GetParent() *Node {
	return (*Node)(n.Parent)
}

func (n *Node) Text() string {
	return n.InnerText()
}

func (n *Node) String() string {
	return n.OutputHTML(true)
}

// Find is like QueryAll but Will panics if the expression `expr` cannot be parsed.
//
// See `QueryAll()` function.
func (n *Node) Find(expr string) []*Node {
	nodes, err := n.QueryAll(expr)
	if err != nil {
		panic(err)
	}
	return nodes
}

// FindOne is like Query but will panics if the expression `expr` cannot be parsed.
// See `Query()` function.
func (n *Node) FindOne(expr string) *Node {
	node, err := n.Query(expr)
	if err != nil {
		panic(err)
	}
	return node
}

// QueryAll searches the html.Node that matches by the specified XPath expr.
// Return an error if the expression `expr` cannot be parsed.
func (n *Node) QueryAll(expr string) ([]*Node, error) {
	exp, err := getQuery(expr)
	if err != nil {
		return nil, err
	}
	nodes := n.QuerySelectorAll(exp)
	return nodes, nil
}

// Query searches the html.Node that matches by the specified XPath expr,
// and return the first element of matched html.Node.
//
// Return an error if the expression `expr` cannot be parsed.
func (n *Node) Query(expr string) (*Node, error) {
	exp, err := getQuery(expr)
	if err != nil {
		return nil, err
	}
	return n.QuerySelector(exp), nil
}

// QuerySelector returns the first matched html.Node by the specified XPath selector.
func (n *Node) QuerySelector(selector *xpath.Expr) *Node {
	t := selector.Select(n.CreateXPathNavigator())
	if t.MoveNext() {
		return getCurrentNode(t.Current().(*NodeNavigator))
	}
	return nil
}

// QuerySelectorAll searches all of the html.Node that matches the specified XPath selectors.
func (n *Node) QuerySelectorAll(selector *xpath.Expr) []*Node {
	var elems []*Node
	t := selector.Select(n.CreateXPathNavigator())
	for t.MoveNext() {
		nav := t.Current().(*NodeNavigator)
		n := getCurrentNode(nav)
		// avoid adding duplicate nodes.
		if len(elems) > 0 && (elems[0] == n || (nav.NodeType() == xpath.AttributeNode &&
			nav.LocalName() == elems[0].Data && nav.Value() == elems[0].InnerText())) {
			continue
		}
		elems = append(elems, n)
	}
	return elems
}

// OutputHTML returns the text including tags name.
func (n *Node) OutputHTML(self bool) string {
	var buf bytes.Buffer

	hn := (*html.Node)(n)
	if self {
		html.Render(&buf, hn)
	} else {
		for n := n.FirstChild; n != nil; n = n.NextSibling {
			html.Render(&buf, hn)
		}
	}
	return buf.String()
}

func (top *Node) InnerText() string {
	n := (*html.Node)(top)
	var output func(*bytes.Buffer, *html.Node)
	output = func(buf *bytes.Buffer, n *html.Node) {
		switch n.Type {
		case html.TextNode:
			buf.WriteString(n.Data)
			return
		case html.CommentNode:
			return
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			output(buf, child)
		}
	}

	var buf bytes.Buffer
	output(&buf, n)
	return buf.String()
}

// CreateXPathNavigator creates a new xpath.NodeNavigator for the specified html.Node.
func (top *Node) CreateXPathNavigator() *NodeNavigator {
	n := (*html.Node)(top)
	return &NodeNavigator{curr: n, root: n, attr: -1}
}
