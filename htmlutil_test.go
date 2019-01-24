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
	"io"
	"strings"
	"testing"
)

func parse(s string, filters ...func(node Node) bool) Node {
	v, err := Parse(strings.NewReader(s), filters...)
	if err != nil {
		panic(err)
	}
	return v
}

func parseElement(s string) Node {
	return parse(
		s,
		func(node Node) bool {
			return node.Data.Type == html.ElementNode && node.Data.Data == "body"
		},
		func(node Node) bool {
			return node.Data.Type == html.ElementNode
		},
	)
}

func TestFilterNodes(t *testing.T) {
	type TestCase struct {
		Input   string
		Filters []func(node Node) bool
		Output  []string
	}
	testCases := []TestCase{
		{
			Input:   "<img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/>",
			Filters: []func(node Node) bool{},
			Output: []string{
				`<html><head></head><body><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></body></html>`,
			},
		},
		{
			Input:   "<img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/>",
			Filters: []func(node Node) bool{nil},
			Output: []string{
				`<html><head></head><body><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></body></html>`,
			},
		},
		{
			Input: "<img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/>",
			Filters: []func(node Node) bool{
				func(node Node) bool {
					return node.Data.Type == html.ElementNode
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
			Filters: []func(node Node) bool{
				func(node Node) bool {
					if node.Data.Data != "head" && node.Data.Data != "body" {
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
			Filters: []func(node Node) bool{
				nil,
				func(node Node) bool {
					if node.Data.Type != html.ElementNode {
						return false
					}
					if node.Data.Data != "div" {
						return false
					}
					if getAttrVal("", "class", node.Data.Attr...) != "one" {
						return false
					}
					return true
				},
				nil,
				nil,
				func(node Node) bool {
					if node.Data.Type != html.ElementNode {
						return false
					}
					if node.Data.Data != "img" {
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
			Filters: []func(node Node) bool{
				func(node Node) bool {
					if node.Data.Type != html.ElementNode {
						return false
					}
					if _, ok := getAttr("", "class", node.Data.Attr...); ok {
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
		for i, v := range filterNodesRaw(input, testCase.Filters...) {
			buffer := new(bytes.Buffer)
			if err := html.Render(buffer, v.Data); err != nil {
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
		Filters []func(node Node) bool
		Output  []string
	}
	testCases := []TestCase{
		{
			Input:   "<img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/>",
			Filters: []func(node Node) bool{},
			Output: []string{
				`<html><head></head><body><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></body></html>`,
			},
		},
		{
			Input:   "<img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/>",
			Filters: []func(node Node) bool{nil},
			Output: []string{
				`<html><head></head><body><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></body></html>`,
			},
		},
		{
			Input: "<img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/>",
			Filters: []func(node Node) bool{
				func(node Node) bool {
					return node.Data.Type == html.ElementNode
				},
			},
			Output: []string{
				`<html><head></head><body><img class="iconClass1" src="/images/icon_1.png" alt="Some Alt Text"/></body></html>`,
			},
		},
		{
			Input: "<img class=\"iconClass1\" src=\"/images/icon_1.png\" alt=\"Some Alt Text\"/>",
			Filters: []func(node Node) bool{
				func(node Node) bool {
					if node.Data.Data != "head" && node.Data.Data != "body" {
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
			Filters: []func(node Node) bool{
				nil,
				func(node Node) bool {
					if node.Data.Type != html.ElementNode {
						return false
					}
					if node.Data.Data != "div" {
						return false
					}
					if getAttrVal("", "class", node.Data.Attr...) != "one" {
						return false
					}
					return true
				},
				nil,
				nil,
				func(node Node) bool {
					if node.Data.Type != html.ElementNode {
						return false
					}
					if node.Data.Data != "img" {
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
			Filters: []func(node Node) bool{
				func(node Node) bool {
					if node.Data.Type != html.ElementNode {
						return false
					}
					if _, ok := getAttr("", "class", node.Data.Attr...); ok {
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
			Filters: []func(node Node) bool{
				func(node Node) bool {
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
		if v, ok := findNodeRaw(input, testCase.Filters...); ok {
			buffer := new(bytes.Buffer)
			if err := html.Render(buffer, v.Data); err != nil {
				t.Fatal(name, i, err)
			}
			output = append(output, buffer.String())
		}
		if diff := deep.Equal(
			len(output),
			len(
				filterNodesWithConfig(
					filterNodesConfig{
						Node:    Node{Data: input},
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
	if v := filterNodesRaw(nil); v != nil {
		t.Fatal(v)
	}
}

func TestEncodeHTML_nil(t *testing.T) {
	if v := encodeHTML(nil); v != "" {
		t.Fatal(v)
	}
}

func TestEncodeHTML_panic(t *testing.T) {
	defer func() {
		if v := fmt.Sprint(recover()); v != "html: cannot render an ErrorNode node" {
			t.Fatal(v)
		}
	}()
	if v := encodeHTML(new(html.Node)); v != "" {
		t.Fatal(v)
	}
}

func TestEncodeText_nil(t *testing.T) {
	if v := encodeText(nil); v != "" {
		t.Fatal(v)
	}
}

func TestParse_eof(t *testing.T) {
	reader, _ := io.Pipe()
	_ = reader.Close()
	if _, err := Parse(reader); err == nil || err.Error() != "io: read/write on closed pipe" {
		t.Fatal(err)
	}
}

func TestParse_notFound(t *testing.T) {
	_, err := Parse(
		strings.NewReader(`<div></div>`),
		func(node Node) bool {
			return false
		},
	)
	if err == nil || err.Error() != "htmlutil.Parse no match" {
		t.Fatal(err)
	}
}

func TestGetNode_nil(t *testing.T) {
	if v := getNodeRaw(nil); v != (Node{}) {
		t.Error(v)
	}
}

func TestGetNode_success(t *testing.T) {
	if v := getNodeRaw(
		parse(`<div><a>b</a><b>a</b></div>`).Data,
		func(node Node) bool {
			return node.Tag() == `b`
		},
	).EncodeHTML(); v != `<b>a</b>` {
		t.Error(v)
	}
}

func TestNode_InnerHTML(t *testing.T) {
	input := `<div>

<div><div>ONE</div><div>TWO</div></div><div>

THREE


</div>FOUR   !

</div>`
	node, err := Parse(
		strings.NewReader(input),
		func(node Node) bool {
			return node.Data.Type == html.ElementNode && node.Data.Data == "div"
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if v := node.InnerHTML(); v != `

<div><div>ONE</div><div>TWO</div></div><div>

THREE


</div>FOUR   !

` {
		t.Fatal(v)
	}
	if v := node.InnerText(); v != `

ONETWO

THREE


FOUR   !

` {
		t.Fatal(v)
	}
}

func TestNode_GetAttr_caseInsensitive(t *testing.T) {
	node := parseElement(`<div one="value_1" TWO="VALUE_2"></div>`)
	if diff := deep.Equal(node.GetAttrVal(``, `one`), `value_1`); diff != nil {
		t.Error(strings.Join(append([]string{"lowercase via lowercase diff:"}, diff...), "    \n"))
	}
	if diff := deep.Equal(node.GetAttrVal(``, `ONE`), `value_1`); diff != nil {
		t.Error(strings.Join(append([]string{"lowercase via uppercase diff:"}, diff...), "    \n"))
	}
	if diff := deep.Equal(node.GetAttrVal(``, `tWo`), `VALUE_2`); diff != nil {
		t.Error(strings.Join(append([]string{"uppercase via mixed diff:"}, diff...), "    \n"))
	}
}

func TestNode_GetAttr_foreign(t *testing.T) {
	node := parseElement(`<svg viewBox="0 0 100 100" xlink:href="#icon-1" not="value"></svg>`)
	if diff := deep.Equal(node.GetAttrVal(`xlink`, `href`), `#icon-1`); diff != nil {
		t.Error(strings.Join(append([]string{"match diff:"}, diff...), "    \n"))
	}
	if diff := deep.Equal(node.GetAttrVal(`XLINK`, `href`), ``); diff != nil {
		t.Error(strings.Join(append([]string{"upper namespace diff:"}, diff...), "    \n"))
	}
	if diff := deep.Equal(node.GetAttrVal(`xlink`, `HREF`), ``); diff != nil {
		t.Error(strings.Join(append([]string{"upper key diff:"}, diff...), "    \n"))
	}
	if diff := deep.Equal(node.GetAttrVal(``, `href`), ``); diff != nil {
		t.Error(strings.Join(append([]string{"empty namespace diff:"}, diff...), "    \n"))
	}
}

func TestNode_FilterNodes_depth(t *testing.T) {
	nodes := parse(
		`<div>
	<div a="1">one</div>
	<div a="2">two</div><div a="3">
		three
		<div a="4">four</div>
	</div>
	<div a="5">five<div a="6">six</div></div></div>`,
	).
		FilterNodes(
			func(node Node) bool {
				return node.Tag() == "div" && node.GetAttrVal(``, `a`) == ""
			},
			func(node Node) bool {
				return node.Type() == html.ElementNode && node.MatchDepth() == 2
			},
		)
	if len(nodes) != 2 {
		t.Fatal(len(nodes))
	}
	if v := nodes[0].Depth; v != 5 {
		t.Error(v)
	}
	if v := nodes[0].MatchDepth(); v != 2 {
		t.Error(v)
	}
	if v := nodes[0].Match; v == nil {
		t.Error(v)
	} else {
		if v := v.Depth; v != 3 {
			t.Error(v)
		}
		if v := v.MatchDepth(); v != 3 {
			t.Error(v)
		}
		if v := v.Match; v != nil {
			t.Error(v)
		}
		if a := nodes[0].Parent().Parent(); a.Data == nil || a.Match != v || a.Match == v.Match {
			t.Error(a.Data)
		} else {
			if v := a.LastChild(); v.Depth != 4 {
				t.Error(v)
			} else if v.Match != a.Match {
				t.Error(v)
			} else if v := v.EncodeHTML(); v != `<div a="5">five<div a="6">six</div></div>` {
				t.Error(v)
			}
			a.Match = v.Match
			if a != *v {
				t.Error(a, "||||||", *v)
			}
		}
	}
	if v := nodes[0].Parent().PrevSibling(); v.Data == nil || v.Depth != 4 {
		t.Error(v)
	} else {
		if v := v.EncodeHTML(); v != `<div a="2">two</div>` {
			t.Error(v)
		}
	}
	if v, ok := nodes[0].GetAttr(``, `A`); !ok || v.Val != "4" {
		t.Error(v, ok)
	}
	if v := nodes[0].Parent(); v.Data == nil {
		t.Error(v)
	} else {
		if v := v.GetAttrVal(``, `A`); v != "3" {
			t.Error(v)
		}
		if v := v.Depth; v != 4 {
			t.Error(v)
		}
	}
	if v := nodes[0].EncodeHTML(); v != `<div a="4">four</div>` {
		t.Error(v)
	}
	if v := nodes[1].EncodeHTML(); v != `<div a="6">six</div>` {
		t.Error(v)
	}
}

func TestNode_Type_nil(t *testing.T) {
	var n Node
	if v := n.Type(); v != 0 {
		t.Fatal(v)
	}
}

func TestNode_Attr_nil(t *testing.T) {
	var n Node
	if v := n.Attr(); v != nil {
		t.Fatal(v)
	}
}

func TestNode_FindNode_success(t *testing.T) {
	n, ok := parse(`<a><b></b><b><c id="search"></c></b></a>`).
		FindNode(
			func(node Node) bool {
				return node.GetAttrVal(``, `id`) == "search"
			},
		)
	if !ok || n.String() != `<c id="search"></c>` {
		t.Fatal(n, ok)
	}
}

func TestNode_Children(t *testing.T) {
	nodes := parseElement(`<a><b></b> <b><c id="search"></c></b></a>`).Children()
	if v := len(nodes); v != 3 {
		t.Fatal(v)
	}
	if v := nodes[0].Type(); v != html.ElementNode {
		t.Fatal(v)
	}
	if v := nodes[0].EncodeHTML(); v != `<b></b>` {
		t.Fatal(v)
	}
	if v := nodes[1].Type(); v != html.TextNode {
		t.Fatal(v)
	}
	if v := nodes[1].EncodeHTML(); v != ` ` {
		t.Fatal(v)
	}
	if v := nodes[2].Type(); v != html.ElementNode {
		t.Fatal(v)
	}
	if v := nodes[2].EncodeHTML(); v != `<b><c id="search"></c></b>` {
		t.Fatal(v)
	}
}

func TestNode_Children_nil(t *testing.T) {
	var n Node
	if v := n.Children(); v != nil {
		t.Fatal(v)
	}
}

func TestNode_GetNode_nil(t *testing.T) {
	if v := (Node{Depth: 1}).GetNode(); v != (Node{}) {
		t.Error(v)
	}
}

func TestNode_GetNode_success(t *testing.T) {
	if v := parse(`<div><a>b</a><b>a</b></div>`).
		GetNode(
			func(node Node) bool {
				return node.Tag() == `b`
			},
		).EncodeHTML(); v != `<b>a</b>` {
		t.Error(v)
	}
}
