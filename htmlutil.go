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

func Parse(r io.Reader, filters ...func(node Node) bool) (Node, error) {
	if node, err := html.Parse(r); err != nil {
		return Node{}, err
	} else if node, ok := FindNode(node, filters...); !ok {
		return Node{}, errors.New("htmlutil.Parse no match")
	} else {
		return node, nil
	}
}

func GetAttr(namespace string, key string, attributes ...html.Attribute) (html.Attribute, bool) {
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

func GetAttrVal(namespace string, key string, attributes ...html.Attribute) string {
	result, _ := GetAttr(namespace, key, attributes...)
	return result.Val
}

func FilterNodes(node *html.Node, filters ...func(node Node) bool) []Node {
	return filterNodes(Node{Data: node}, filters...)
}

func FindNode(node *html.Node, filters ...func(node Node) bool) (Node, bool) {
	return findNode(Node{Data: node}, filters...)
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
