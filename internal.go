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
	filterNodesConfig struct {
		Node    Node
		Filters []func(node Node) bool
		Find    bool
	}
)

func copyValidFilters(input []func(node Node) bool) (output []func(node Node) bool) {
	for _, filter := range input {
		if filter != nil {
			output = append(output, filter)
		}
	}
	return
}

func filterNodesWithConfig(config filterNodesConfig) []Node {
	config.Filters = copyValidFilters(config.Filters)

	var (
		result []Node
		fn     func(config filterNodesConfig)
	)

	fn = func(config filterNodesConfig) {
		if config.Node.Data == nil {
			return
		}

		if config.Node.Match == nil {
			match := config.Node
			config.Node.Match = &match
		}

		if config.Find && len(result) != 0 {
			return
		}

		if len(config.Filters) == 0 {
			result = append(result, config.Node)
			return
		}

		start := len(result)

		func(config filterNodesConfig) {
			var filter func(node Node) bool

			for filter == nil && len(config.Filters) != 0 {
				filter = config.Filters[0]
				config.Filters = config.Filters[1:]
			}

			if filter != nil && !filter(config.Node) {
				return
			}

			if len(config.Filters) == 0 {
				fn(config)

				return
			}

			match := config.Node
			config.Node.Match = &match

			for c := config.Node.FirstChild(); c.Data != nil; c = c.NextSibling() {
				config.Node = c

				fn(config)
			}
		}(config)

		finish := len(result)

		for c := config.Node.FirstChild(); c.Data != nil; c = c.NextSibling() {
			config.Node = c

			fn(config)

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

	fn(config)

	return result
}

func filterNodes(node Node, filters ...func(node Node) bool) []Node {
	return filterNodesWithConfig(
		filterNodesConfig{
			Node:    node,
			Filters: filters,
		},
	)
}

func findNode(node Node, filters ...func(node Node) bool) (Node, bool) {
	elements := filterNodesWithConfig(
		filterNodesConfig{
			Node:    node,
			Filters: filters,
			Find:    true,
		},
	)
	if len(elements) == 0 {
		return Node{}, false
	}
	return elements[0], true
}

func getNode(node Node, filters ...func(node Node) bool) Node {
	result, _ := findNode(node, filters...)
	return result
}

func encodeTextBytes(node *html.Node) []byte {
	if node == nil {
		return nil
	}
	if node.Type == html.TextNode {
		return []byte(node.Data)
	}
	var b []byte
	for node := node.FirstChild; node != nil; node = node.NextSibling {
		b = append(b, encodeTextBytes(node)...)
	}
	return b
}

func encodeText(node *html.Node) string {
	return string(encodeTextBytes(node))
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

func filterNodesRaw(node *html.Node, filters ...func(node Node) bool) []Node {
	return filterNodes(Node{Data: node}, filters...)
}

func findNodeRaw(node *html.Node, filters ...func(node Node) bool) (Node, bool) {
	return findNode(Node{Data: node}, filters...)
}

func getNodeRaw(node *html.Node, filters ...func(node Node) bool) Node {
	return getNode(Node{Data: node}, filters...)
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
