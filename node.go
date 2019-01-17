package htmlutil

import (
	"bytes"
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

func (n Node) String() string {
	if n.Data == nil {
		return ""
	}
	buffer := new(bytes.Buffer)
	if err := html.Render(buffer, n.Data); err != nil {
		panic(err)
	}
	return buffer.String()
}
