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
	"errors"
	"golang.org/x/net/html"
	"io"
	"strings"
)

func Parse(r io.Reader, filters ...func(node *html.Node) bool) (Node, error) {
	node, err := html.Parse(r)
	if err != nil {
		return Node{}, err
	}
	node, ok := FindNode(node, filters...)
	if !ok {
		return Node{}, errors.New("htmlutil.Parse no match")
	}
	return Node{
		Data: node,
	}, nil
}

func GetAttribute(namespace string, key string, attributes ...html.Attribute) (html.Attribute, bool) {
	keyCaseInsensitive := namespace == ``
	if keyCaseInsensitive {
		key = strings.ToLower(key)
	}

	for _, attribute := range attributes {
		if attribute.Namespace != namespace {
			continue
		}

		if keyCaseInsensitive {
			if strings.ToLower(attribute.Key) != key {
				continue
			}
		} else if attribute.Key != key {
			continue
		}

		return attribute, true
	}

	return html.Attribute{}, false
}

func GetAttributeValue(namespace string, key string, attributes ...html.Attribute) string {
	result, _ := GetAttribute(namespace, key, attributes...)
	return result.Val
}

func FilterNodes(node *html.Node, filters ...func(node *html.Node) bool) []*html.Node {
	return filterNodes(
		filterNodesConfig{
			Node:    node,
			Filters: filters,
		},
	)
}

func FindNode(node *html.Node, filters ...func(node *html.Node) bool) (*html.Node, bool) {
	elements := filterNodes(
		filterNodesConfig{
			Node:    node,
			Filters: filters,
			Find:    true,
		},
	)
	if len(elements) == 0 {
		return nil, false
	}
	return elements[0], true
}

func EncodeHTML(node *html.Node) string {
	if node == nil {
		return ""
	}
	buffer := new(bytes.Buffer)
	if err := html.Render(buffer, node); err != nil {
		panic(err)
	}
	return buffer.String()
}

func EncodeText(node *html.Node) string {
	return string(encodeText(node))
}
