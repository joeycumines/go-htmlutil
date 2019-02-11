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

// Package htmlutil implements a wrapper for Golang's html5 tokeniser / parser implementation, making it much easier to
// find and extract information, aiming to be powerful and intuitive while remaining a minimal and logical extension.
//
// There are three core components, the `htmlutil.Node` struct (a wrapper for `*html.Node`), the `htmlutil.Parse`
// function (optional), an ubiquitous filter algorithm used throughout this implementation, providing functionality
// similar to CSS selectors, and powered by optional (varargs) parameters in the form of chained closures with a
// signature of `func(htmlutil.Node) bool`.
//
// Filter behavior
//
// - based on a recursive algorithm where each node can match at most one filter, consuming it (for that sub-tree),
//   and is added to the result if `len(filters) == 0`
// - every node in the tree is searched (in general, there is a "find" mode where only one result is returned)
// - nil filters are preemptively stripped, and so are treated like they were omitted
// - each node will be present in the result at most once, and will retain (depth first) order
// - behavior is undefined if the tree is not "well formed" (e.g. any cycles)
// - providing no filters will return ALL nodes (or if only one result is needed, the first node)
// - filter closures will not be called with a node with a nil `Data` field
// - filter closures will receive nodes with a `Depth` field relative to the original
// - the node's `Match` field stores the last "matched" node in the chain (note: duplicate matches for the same
//   `*html.Node` are squashed), the root node is always treated as an initial match
// - resulting node values will retain the match chain (will always be non-nil if the root was non-nil)
//
// General behavior
//
// - a nil `Data` field for a `htmlutil.Node` indicates no node / no result, and methods should return default values,
//   or other intuitive analog (behavior to make chaining far simpler)
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

// Parse first performs html.Parse, parsing through any errors, before applying a find to the resulting Node (wrapped
// like `Node{Data: node}`), returning the first matching Node, or an error, if no matches were found
func Parse(r io.Reader, filters ...func(node Node) bool) (Node, error) {
	if node, err := html.Parse(r); err != nil {
		return Node{}, err
	} else if node, ok := findNode(Node{Data: node}, filters...); !ok {
		return Node{}, errors.New("htmlutil.Parse no match")
	} else {
		return node, nil
	}
}

// Attr will return the value of `n.Data.Attr`, returning nil if `n.Data` is nil
func (n Node) Attr() []html.Attribute {
	if n.Data == nil {
		return nil
	}
	return n.Data.Attr
}

// Offset is the difference between the depth of this node and the depth of last match, returning the depth of this
// node if `n.Match` is nil
func (n Node) Offset() int {
	d := n.Depth
	if n.Match != nil {
		d -= n.Match.Depth
	}
	return d
}

// Type will return the value of `n.Data.Type`, returning `html.ErrorNode` if `n.Data` is nil
func (n Node) Type() html.NodeType {
	if n.Data != nil {
		return n.Data.Type
	}
	return html.ErrorNode
}

// Tag will return `n.Data.Data` if the node has a type of `html.ElementNode`, otherwise it will return an empty string
func (n Node) Tag() string {
	if n.Type() == html.ElementNode {
		return n.Data.Data
	}
	return ""
}

// GetAttr matches on the first attribute (if any) for this node with the same namespace and key (key being case
// insensitive if namespace is empty), returning false if no match was found
func (n Node) GetAttr(namespace string, key string) (html.Attribute, bool) {
	return getAttr(namespace, key, n.Attr()...)
}

// GetAttrVal returns the value of any attribute matched by `n.GetAttr`
func (n Node) GetAttrVal(namespace string, key string) string {
	return getAttrVal(namespace, key, n.Attr()...)
}

// String is an alias for `n.OuterHTML`
func (n Node) String() string {
	return n.OuterHTML()
}

// OuterHTML encodes this node as html using the `html.Render` function, note that it will return an empty string
// if `n.Data` is nil, and will panic if any error is returned (which should only occur if the sub-tree is not
// "well formed")
func (n Node) OuterHTML() string {
	return encodeHTML(n.Data)
}

// OuterText builds a string from the data of all text nodes in the sub-tree, starting from and including `n`
func (n Node) OuterText() string {
	return encodeText(n.Data)
}

// InnerHTML builds a string using the outer html of all children matching all filters (see the `FindNode` method)
func (n Node) InnerHTML(filters ...func(node Node) bool) string {
	var b []byte
	n.Range(
		func(i int, node Node) bool {
			b = append(b, []byte(node.OuterHTML())...)
			return true
		},
		filters...,
	)
	return string(b)
}

// InnerText builds a string using the outer text of all children matching all filters (see the `FindNode` method)
func (n Node) InnerText(filters ...func(node Node) bool) string {
	var b []byte
	n.Range(
		func(i int, node Node) bool {
			b = append(b, []byte(node.OuterText())...)
			return true
		},
		filters...,
	)
	return string(b)
}

// SiblingIndex returns the total number of previous siblings matching any filters (see the `FindNode` method)
func (n Node) SiblingIndex(filters ...func(node Node) bool) int {
	return siblingIndex(n, filters...)
}

// SiblingLength returns the total number of siblings matching any filters (see the `FindNode` method) incremented by
// one for the current node, or returns 0 if the receiver has nil data (is empty)
func (n Node) SiblingLength(filters ...func(node Node) bool) int {
	return siblingLength(n, filters...)
}

// FilterNodes returns all nodes from the sub-tree (a search including the receiver) matching the filters (see package
// comment for filter behavior)
func (n Node) FilterNodes(filters ...func(node Node) bool) []Node {
	return filterNodes(n, filters...)
}

// FindNode returns the first node from the sub-tree (a search including the receiver) matching the filters (see
// package comment for filter behavior)
func (n Node) FindNode(filters ...func(node Node) bool) (Node, bool) {
	return findNode(n, filters...)
}

// GetNode returns the node returned by FindNode without the boolean flag indicating if there was a match, it is
// provided for chaining purposes, since this package deliberately handles a nil `Data` field gracefully
func (n Node) GetNode(filters ...func(node Node) bool) Node {
	return getNode(n, filters...)
}

// Range iterates on any children matching any filters (see the `FindNode` method), providing the (filtered) index
// and node to the provided fn, note that it will panic if fn is nil
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

// Children builds a slice containing all child nodes using the `Range` method, passing through filters
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

// Parent will return the first parent node matching any filters (see the `FindNode` method), or a node with a nil
// `Data` property for no match, note that depth will be automatically decremented (potentially multiple times)
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

// FirstChild will return the leftmost child node matching any filters (see the `FindNode` method), or a node with a
// nil `Data` property for no match, note that depth will be automatically incremented
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

// LastChild will return the rightmost child node matching any filters (see the `FindNode` method), or a node with a
// nil `Data` property for no match, note that depth will be automatically incremented
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

// PrevSibling will return the rightmost previous sibling node matching any filters (see the `FindNode` method), or a
// node with a nil `Data` property for no match
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

// NextSibling will return the leftmost next sibling node matching any filters (see the `FindNode` method), or a
// node with a nil `Data` property for no match
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
