package htmlutil

import (
	"errors"
	"golang.org/x/net/html"
	"io"
)

type Node struct {
	Data *html.Node
}

func Parse(r io.Reader, filters ...func(node *html.Node) bool) (Node, error) {
	node, err := html.Parse(r)
	if err != nil {
		return Node{}, err
	}
	node, ok := FindElement(node, filters...)
	if !ok {
		return Node{}, errors.New("htmlutil.Parse no match")
	}
	return Node{
		Data: node,
	}, nil
}

func (n Node) FilterElements(filters ...func(node *html.Node) bool) []Node {
	nodes := FilterElements(n.Data, filters...)
	if nodes == nil {
		return nil
	}
	result := make([]Node, len(nodes))
	for i, node := range nodes {
		result[i] = Node{
			Data: node,
		}
	}
	return result
}

func (n Node) FindElement(filters ...func(node *html.Node) bool) (Node, bool) {
	node, ok := FindElement(n.Data, filters...)
	return Node{Data: node}, ok
}

func (n Node) Attributes() []html.Attribute {
	if n.Data == nil {
		return nil
	}
	return n.Data.Attr
}

func (n Node) GetAttribute(namespace string, key string) (html.Attribute, bool) {
	return GetAttribute(namespace, key, n.Attributes()...)
}

func (n Node) GetAttributeValue(namespace string, key string, attributes ...html.Attribute) string {
	return GetAttributeValue(namespace, key, n.Attributes()...)
}

func (n Node) EncodeHTML() string {
	return EncodeHTML(n.Data)
}

func (n Node) EncodeText() string {
	return EncodeText(n.Data)
}

func (n Node) String() string {
	return n.EncodeHTML()
}

func (n Node) Children() (children []Node) {
	if n.Data == nil {
		return
	}
	for c := n.Data.FirstChild; c != nil; c = c.NextSibling {
		children = append(children, Node{Data: c})
	}
	return
}

func (n Node) InnerHTML() string {
	var b []byte
	for _, child := range n.Children() {
		b = append(b, []byte(child.EncodeHTML())...)
	}
	return string(b)
}

func (n Node) InnerText() string {
	var b []byte
	for _, child := range n.Children() {
		b = append(b, []byte(child.EncodeText())...)
	}
	return string(b)
}
