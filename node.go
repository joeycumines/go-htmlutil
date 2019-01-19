/*
   Copyright 2019 Joseph Cumines

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */

package htmlutil

import (
	"golang.org/x/net/html"
)

type Node struct {
	Data *html.Node
}

func (n Node) FilterNodes(filters ...func(node *html.Node) bool) []Node {
	nodes := FilterNodes(n.Data, filters...)
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

func (n Node) FindNode(filters ...func(node *html.Node) bool) (Node, bool) {
	node, ok := FindNode(n.Data, filters...)
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
