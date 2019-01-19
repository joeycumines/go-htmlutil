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
	"fmt"
	"github.com/go-test/deep"
	"golang.org/x/net/html"
	"strings"
	"testing"
)

func parse(s string, filters ...func(node *html.Node) bool) Node {
	v, err := Parse(strings.NewReader(s), filters...)
	if err != nil {
		panic(err)
	}
	return v
}

func parseElement(s string) Node {
	return parse(
		s,
		func(node *html.Node) bool {
			return node.Type == html.ElementNode && node.Data == "body"
		},
		func(node *html.Node) bool {
			return node.Type == html.ElementNode
		},
	)
}

func TestFilterNodes(t *testing.T) {
	type TestCase struct {
		Input   string
		Filters []func(node *html.Node) bool
		Output  []string
	}
	testCases := []TestCase{
		{
			Input:   "<img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/>",
			Filters: []func(node *html.Node) bool{},
			Output: []string{
				`<html><head></head><body><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></body></html>`,
			},
		},
		{
			Input:   "<img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/>",
			Filters: []func(node *html.Node) bool{nil},
			Output: []string{
				`<html><head></head><body><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></body></html>`,
			},
		},
		{
			Input: "<img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/>",
			Filters: []func(node *html.Node) bool{
				func(node *html.Node) bool {
					return node.Type == html.ElementNode
				},
			},
			Output: []string{
				`<html><head></head><body><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></body></html>`,
				`<head></head>`,
				`<body><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></body>`,
				`<img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/>`,
			},
		},
		{
			Input: "<img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/>",
			Filters: []func(node *html.Node) bool{
				func(node *html.Node) bool {
					if node.Data != "head" && node.Data != "body" {
						return false
					}
					return true
				},
			},
			Output: []string{
				`<head></head>`,
				`<body><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></body>`,
			},
		},
		{
			Input: "<div class=\"one\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"top level\"/><div class=\"one\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"further nested\"/></div></div><div class=\"one\"></div><div class=\"two\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/></div><div class=\"one\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"final\"/></div>",
			Filters: []func(node *html.Node) bool{
				nil,
				func(node *html.Node) bool {
					if node.Type != html.ElementNode {
						return false
					}
					if node.Data != "div" {
						return false
					}
					if GetAttributeValue("", "class", node.Attr...) != "one" {
						return false
					}
					return true
				},
				nil,
				nil,
				func(node *html.Node) bool {
					if node.Type != html.ElementNode {
						return false
					}
					if node.Data != "img" {
						return false
					}
					return true
				},
				nil,
			},
			Output: []string{
				`<img class="iconClass1" src="/images/icon_1.png" alt="top level"/>`,
				`<img class="iconClass1" src="/images/icon_1.png" alt="further nested"/>`,
				`<img class="iconClass1" src="/images/icon_1.png" alt="final"/>`,
			},
		},
		{
			Input: "<div class=\"one\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"top level\"/><div class=\"one\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"further nested\"/></div></div><div class=\"one\"></div><div class=\"two\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/></div><div class=\"one\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"final\"/></div>",
			Filters: []func(node *html.Node) bool{
				func(node *html.Node) bool {
					if node.Type != html.ElementNode {
						return false
					}
					if _, ok := GetAttribute("", "class", node.Attr...); ok {
						return false
					}
					return true
				},
			},
			Output: []string{
				`<html><head></head><body><div class="one"><img class="iconClass1" src="/images/icon_1.png" alt="top level"/><div class="one"><img class="iconClass1" src="/images/icon_1.png" alt="further nested"/></div></div><div class="one"></div><div class="two"><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></div><div class="one"><img class="iconClass1" src="/images/icon_1.png" alt="final"/></div></body></html>`,
				`<head></head>`,
				`<body><div class="one"><img class="iconClass1" src="/images/icon_1.png" alt="top level"/><div class="one"><img class="iconClass1" src="/images/icon_1.png" alt="further nested"/></div></div><div class="one"></div><div class="two"><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></div><div class="one"><img class="iconClass1" src="/images/icon_1.png" alt="final"/></div></body>`,
			},
		},
	}
	for i, testCase := range testCases {
		name := fmt.Sprintf("FilterNodes_#%d", i+1)
		input, err := html.Parse(bytes.NewReader([]byte(testCase.Input)))
		if err != nil {
			t.Fatal(name, err)
		}
		var output []string
		for i, v := range FilterNodes(input, testCase.Filters...) {
			buffer := new(bytes.Buffer)
			if err := html.Render(buffer, v); err != nil {
				t.Fatal(name, i, err)
			}
			output = append(output, buffer.String())
		}
		if diff := deep.Equal(
			output,
			testCase.Output,
		); diff != nil {
			t.Error(strings.Join(append([]string{name + " output diff:"}, diff...), "    \n"))
		}
	}
}

func TestFindNode(t *testing.T) {
	type TestCase struct {
		Input   string
		Filters []func(node *html.Node) bool
		Output  []string
	}
	testCases := []TestCase{
		{
			Input:   "<img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/>",
			Filters: []func(node *html.Node) bool{},
			Output: []string{
				`<html><head></head><body><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></body></html>`,
			},
		},
		{
			Input:   "<img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/>",
			Filters: []func(node *html.Node) bool{nil},
			Output: []string{
				`<html><head></head><body><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></body></html>`,
			},
		},
		{
			Input: "<img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/>",
			Filters: []func(node *html.Node) bool{
				func(node *html.Node) bool {
					return node.Type == html.ElementNode
				},
			},
			Output: []string{
				`<html><head></head><body><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></body></html>`,
			},
		},
		{
			Input: "<img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/>",
			Filters: []func(node *html.Node) bool{
				func(node *html.Node) bool {
					if node.Data != "head" && node.Data != "body" {
						return false
					}
					return true
				},
			},
			Output: []string{
				`<head></head>`,
			},
		},
		{
			Input: "<div class=\"one\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"top level\"/><div class=\"one\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"further nested\"/></div></div><div class=\"one\"></div><div class=\"two\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/></div><div class=\"one\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"final\"/></div>",
			Filters: []func(node *html.Node) bool{
				nil,
				func(node *html.Node) bool {
					if node.Type != html.ElementNode {
						return false
					}
					if node.Data != "div" {
						return false
					}
					if GetAttributeValue("", "class", node.Attr...) != "one" {
						return false
					}
					return true
				},
				nil,
				nil,
				func(node *html.Node) bool {
					if node.Type != html.ElementNode {
						return false
					}
					if node.Data != "img" {
						return false
					}
					return true
				},
				nil,
			},
			Output: []string{
				`<img class="iconClass1" src="/images/icon_1.png" alt="top level"/>`,
			},
		},
		{
			Input: "<div class=\"one\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"top level\"/><div class=\"one\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"further nested\"/></div></div><div class=\"one\"></div><div class=\"two\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/></div><div class=\"one\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"final\"/></div>",
			Filters: []func(node *html.Node) bool{
				func(node *html.Node) bool {
					if node.Type != html.ElementNode {
						return false
					}
					if _, ok := GetAttribute("", "class", node.Attr...); ok {
						return false
					}
					return true
				},
			},
			Output: []string{
				`<html><head></head><body><div class="one"><img class="iconClass1" src="/images/icon_1.png" alt="top level"/><div class="one"><img class="iconClass1" src="/images/icon_1.png" alt="further nested"/></div></div><div class="one"></div><div class="two"><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></div><div class="one"><img class="iconClass1" src="/images/icon_1.png" alt="final"/></div></body></html>`,
			},
		},
		{
			Input: "<div class=\"one\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"top level\"/><div class=\"one\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"further nested\"/></div></div><div class=\"one\"></div><div class=\"two\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/></div><div class=\"one\"><img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"final\"/></div>",
			Filters: []func(node *html.Node) bool{
				func(node *html.Node) bool {
					return false
				},
			},
			Output: nil,
		},
	}
	for i, testCase := range testCases {
		name := fmt.Sprintf("FindNode_#%d", i+1)
		input, err := html.Parse(bytes.NewReader([]byte(testCase.Input)))
		if err != nil {
			t.Fatal(name, err)
		}
		var output []string
		if v, ok := FindNode(input, testCase.Filters...); ok {
			buffer := new(bytes.Buffer)
			if err := html.Render(buffer, v); err != nil {
				t.Fatal(name, i, err)
			}
			output = append(output, buffer.String())
		}
		if diff := deep.Equal(
			len(output),
			len(
				filterNodes(
					filterNodesConfig{
						Node:    input,
						Filters: testCase.Filters,
						Find:    true,
					},
				),
			),
		); diff != nil {
			t.Error(strings.Join(append([]string{name + " find len diff:"}, diff...), "    \n"))
		}
		if diff := deep.Equal(
			output,
			testCase.Output,
		); diff != nil {
			t.Error(strings.Join(append([]string{name + " output diff:"}, diff...), "    \n"))
		}
	}
}

func TestFilterNodes_nil(t *testing.T) {
	if v := FilterNodes(nil); v != nil {
		t.Fatal(v)
	}
}
