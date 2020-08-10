package htmlquery

import "golang.org/x/net/html"

// Attribute 属性
type Attribute html.Attribute

func (attr *Attribute) GetKey() string {
	return attr.Key
}

func (attr *Attribute) GetValue() string {
	return attr.Val
}

func (attr *Attribute) GetNamespace() string {
	return attr.Namespace
}
