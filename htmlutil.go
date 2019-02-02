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
	"errors"
	"golang.org/x/net/html"
	"io"
)

// Node is the data structure this package provides to allow utilisation of utility methods + extra metadata such
// as the last match (`Match` property) for filter / find / get calls, as well as the overall (relative) depth,
// allowing matching on things such as "all the table row elements that are direct children of a given tbody", a-la
// CSS selectors
type Node struct {
	// Data is the underlying html data for this node
	Data *html.Node
	// Depth is the relative depth to the top of the tree (being parsed, filtered, etc)
	Depth int
	// Match is the last match (set by filter impl.), and is used to check previous matches for chained filters
	Match *Node
}

func Parse(r io.Reader, filters ...func(node Node) bool) (Node, error) {
	if node, err := html.Parse(r); err != nil {
		return Node{}, err
	} else if node, ok := findNode(Node{Data: node}, filters...); !ok {
		return Node{}, errors.New("htmlutil.Parse no match")
	} else {
		return node, nil
	}
}

func (n Node) FilterNodes(filters ...func(node Node) bool) []Node {
	return filterNodes(n, filters...)
}

func (n Node) FindNode(filters ...func(node Node) bool) (Node, bool) {
	return findNode(n, filters...)
}

func (n Node) GetNode(filters ...func(node Node) bool) Node {
	return getNode(n, filters...)
}

func (n Node) Attr() []html.Attribute {
	if n.Data == nil {
		return nil
	}
	return n.Data.Attr
}

func (n Node) GetAttr(namespace string, key string) (html.Attribute, bool) {
	return getAttr(namespace, key, n.Attr()...)
}

func (n Node) GetAttrVal(namespace string, key string) string {
	return getAttrVal(namespace, key, n.Attr()...)
}

func (n Node) EncodeHTML() string {
	return encodeHTML(n.Data)
}

func (n Node) EncodeText() string {
	return encodeText(n.Data)
}

func (n Node) String() string {
	return n.EncodeHTML()
}

func (n Node) Range(fn func(i int, node Node) bool, filters ...func(node Node) bool) {
	if fn == nil {
		panic(errors.New("htmlutil.Node.Range nil fn"))
	}
	i := 0
	for node := n.FirstChild(filters...); node.Data != nil; node = node.NextSibling(filters...) {
		if !fn(i, node) {
			break
		}
		i++
	}
}

func (n Node) Children(filters ...func(node Node) bool) (children []Node) {
	n.Range(
		func(i int, node Node) bool {
			children = append(children, node)
			return true
		},
		filters...,
	)
	return
}

func (n Node) InnerHTML(filters ...func(node Node) bool) string {
	var b []byte
	n.Range(
		func(i int, node Node) bool {
			b = append(b, []byte(node.EncodeHTML())...)
			return true
		},
		filters...,
	)
	return string(b)
}

func (n Node) InnerText(filters ...func(node Node) bool) string {
	var b []byte
	n.Range(
		func(i int, node Node) bool {
			b = append(b, []byte(node.EncodeText())...)
			return true
		},
		filters...,
	)
	return string(b)
}

func (n Node) Parent(filters ...func(node Node) bool) Node {
	n.Depth--
	if n.Data != nil {
		n.Data = n.Data.Parent
	}
	if n.Data != nil {
		if _, ok := n.FindNode(filters...); !ok {
			return n.Parent(filters...)
		}
	}
	return n
}

func (n Node) FirstChild(filters ...func(node Node) bool) Node {
	n.Depth++
	if n.Data != nil {
		n.Data = n.Data.FirstChild
	}
	if n.Data != nil {
		if _, ok := n.FindNode(filters...); !ok {
			return n.NextSibling(filters...)
		}
	}
	return n
}

func (n Node) LastChild(filters ...func(node Node) bool) Node {
	n.Depth++
	if n.Data != nil {
		n.Data = n.Data.LastChild
	}
	if n.Data != nil {
		if _, ok := n.FindNode(filters...); !ok {
			return n.PrevSibling(filters...)
		}
	}
	return n
}

func (n Node) PrevSibling(filters ...func(node Node) bool) Node {
	if n.Data != nil {
		n.Data = n.Data.PrevSibling
	}
	if n.Data != nil {
		if _, ok := n.FindNode(filters...); !ok {
			return n.PrevSibling(filters...)
		}
	}
	return n
}

func (n Node) NextSibling(filters ...func(node Node) bool) Node {
	if n.Data != nil {
		n.Data = n.Data.NextSibling
	}
	if n.Data != nil {
		if _, ok := n.FindNode(filters...); !ok {
			return n.NextSibling(filters...)
		}
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

// SiblingIndex returns the total number of previous siblings matching any filters
func (n Node) SiblingIndex(filters ...func(node Node) bool) int {
	return siblingIndex(n, filters...)
}

// SiblingLength returns the total number of siblings matching any filters incremented by one for the current node,
// or returns 0 if the receiver has nil data (is empty)
func (n Node) SiblingLength(filters ...func(node Node) bool) int {
	return siblingLength(n, filters...)
}
