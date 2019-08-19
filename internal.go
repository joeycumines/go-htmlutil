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
	"bytes"
	"golang.org/x/net/html"
	"strings"
)

type (
	filterConfig struct {
		Node    Node
		Filters []func(node Node) bool
		Find    bool
	}
)

func (c filterConfig) filters() []func(node Node) bool {
	var result []func(node Node) bool
	for _, filter := range c.Filters {
		if filter != nil {
			result = append(result, filter)
		}
	}
	return result
}

func (c filterConfig) match() *Node {
	if c.Node.Data == nil {
		return c.Node.Match
	}
	if c.Node.Match != nil && c.Node.Match.Data == c.Node.Data {
		return c.Node.Match
	}
	return &c.Node
}

func (c filterConfig) filter() []Node {
	c.Filters = c.filters()

	c.Node.Match = c.match()

	var (
		result []Node
		fn     func(c filterConfig)
	)

	fn = func(c filterConfig) {
		if c.Node.Data == nil {
			return
		}

		if c.Find && len(result) != 0 {
			return
		}

		if len(c.Filters) == 0 {
			result = append(result, c.Node)
			return
		}

		start := len(result)

		func(c filterConfig) {
			var filter func(node Node) bool

			for filter == nil && len(c.Filters) != 0 {
				filter = c.Filters[0]
				c.Filters = c.Filters[1:]
			}

			if filter != nil && !filter(c.Node) {
				return
			}

			if len(c.Filters) == 0 {
				fn(c)

				return
			}

			c.Node.Match = c.match()

			for n := c.Node.FirstChild(); n.Data != nil; n = n.NextSibling() {
				c.Node = n

				fn(c)
			}
		}(c)

		finish := len(result)

		for n := c.Node.FirstChild(); n.Data != nil; n = n.NextSibling() {
			c.Node = n

			fn(c)

			for i := start; i < finish; i++ {
				for j := finish; j < len(result); j++ {
					if result[i].Data != result[j].Data {
						continue
					}

					copy(result[j:], result[j+1:])
					result[len(result)-1] = Node{}
					result = result[:len(result)-1]
					j--
				}
			}
		}
	}

	fn(c)

	return result
}

func filterNodes(node Node, filters ...func(node Node) bool) []Node {
	return (filterConfig{
		Node:    node,
		Filters: filters,
	}).filter()
}

func findNode(node Node, filters ...func(node Node) bool) (Node, bool) {
	elements := (filterConfig{
		Node:    node,
		Filters: filters,
		Find:    true,
	}).filter()
	if len(elements) == 0 {
		return Node{}, false
	}
	return elements[0], true
}

func getNode(node Node, filters ...func(node Node) bool) Node {
	result, _ := findNode(node, filters...)
	return result
}

func encodeHTML(node *html.Node) string {
	if node == nil {
		return ""
	}
	buffer := new(bytes.Buffer)
	if err := html.Render(buffer, node); err != nil {
		panic(err)
	}
	return buffer.String()
}

func encodeText(node *html.Node) []byte {
	if node == nil {
		return nil
	}
	if node.Type == html.TextNode {
		return []byte(node.Data)
	}
	var b []byte
	for node := node.FirstChild; node != nil; node = node.NextSibling {
		b = append(b, encodeText(node)...)
	}
	return b
}

func encodeWords(node *html.Node) (b []byte) {
	if node == nil {
		return
	}
	if node.Type == html.TextNode {
		for _, word := range strings.Fields(node.Data) {
			if len(b) != 0 {
				b = append(b, ' ')
			}
			b = append(b, []byte(word)...)
		}
		return
	}
	for node := node.FirstChild; node != nil; node = node.NextSibling {
		if words := encodeWords(node); len(words) != 0 {
			if len(b) != 0 {
				b = append(b, ' ')
			}
			b = append(b, words...)
		}
	}
	return
}

func getAttr(namespace string, key string, attributes ...html.Attribute) (html.Attribute, bool) {
	keyCaseInsensitive := namespace == ``
	if keyCaseInsensitive {
		key = strings.ToLower(key)
	}

	for _, attr := range attributes {
		if attr.Namespace != namespace {
			continue
		}

		if keyCaseInsensitive {
			if strings.ToLower(attr.Key) != key {
				continue
			}
		} else if attr.Key != key {
			continue
		}

		return attr, true
	}

	return html.Attribute{}, false
}

func getAttrVal(namespace string, key string, attributes ...html.Attribute) string {
	result, _ := getAttr(namespace, key, attributes...)
	return result.Val
}

func siblingIndex(node Node, filters ...func(node Node) bool) (v int) {
	// results the count of previous siblings matching any filters
	for node = node.PrevSibling(filters...); node.Data != nil; node = node.PrevSibling(filters...) {
		v++
	}
	return
}

func siblingLength(node Node, filters ...func(node Node) bool) (v int) {
	// count previous siblings matching filters
	v = siblingIndex(node, filters...)
	// count this node (if not empty) and filtered next siblings
	for ; node.Data != nil; node = node.NextSibling(filters...) {
		v++
	}
	return
}
