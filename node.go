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
	// Data is the underlying html data for this node
	Data *html.Node
	// Depth is the relative depth to the top of the tree (being parsed, filtered, etc)
	Depth int
	// Match is the last match (set by filter impl.), and is used to check previous matches for chained filters
	Match *Node
}

func (n Node) FilterNodes(filters ...func(node Node) bool) []Node {
	return filterNodes(n, filters...)
}

func (n Node) FindNode(filters ...func(node Node) bool) (Node, bool) {
	return findNode(n, filters...)
}

func (n Node) Attr() []html.Attribute {
	if n.Data == nil {
		return nil
	}
	return n.Data.Attr
}

func (n Node) GetAttr(namespace string, key string) (html.Attribute, bool) {
	return GetAttr(namespace, key, n.Attr()...)
}

func (n Node) GetAttrVal(namespace string, key string, attributes ...html.Attribute) string {
	return GetAttrVal(namespace, key, n.Attr()...)
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
	for child := n.FirstChild(); child.Data != nil; child = child.NextSibling() {
		children = append(children, child)
	}
	return
}

func (n Node) InnerHTML() string {
	var b []byte
	for child := n.FirstChild(); child.Data != nil; child = child.NextSibling() {
		b = append(b, []byte(child.EncodeHTML())...)
	}
	return string(b)
}

func (n Node) InnerText() string {
	var b []byte
	for child := n.FirstChild(); child.Data != nil; child = child.NextSibling() {
		b = append(b, []byte(child.EncodeText())...)
	}
	return string(b)
}

func (n Node) Parent() Node {
	if n.Data != nil {
		n.Data = n.Data.Parent
	}
	n.Depth--
	return n
}

func (n Node) FirstChild() Node {
	if n.Data != nil {
		n.Data = n.Data.FirstChild
	}
	n.Depth++
	return n
}

func (n Node) LastChild() Node {
	if n.Data != nil {
		n.Data = n.Data.LastChild
	}
	n.Depth++
	return n
}

func (n Node) PrevSibling() Node {
	if n.Data != nil {
		n.Data = n.Data.PrevSibling
	}
	return n
}

func (n Node) NextSibling() Node {
	if n.Data != nil {
		n.Data = n.Data.NextSibling
	}
	return n
}

func (n Node) MatchDepth() int {
	d := n.Depth
	if n.Match != nil {
		d -= n.Match.Depth
	}
	return d
}

func (n Node) Type() html.NodeType {
	if n.Data != nil {
		return n.Data.Type
	}
	return html.ErrorNode
}

func (n Node) Tag() string {
	if n.Type() == html.ElementNode {
		return n.Data.Data
	}
	return ""
}
