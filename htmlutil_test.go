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
				(filterConfig{
					Node:    Node{Data: input},
					Filters: testCase.Filters,
					Find:    true,
				}).filter(),
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
	if v := encodeText(nil); v != nil {
		t.Fatal(v)
	}
}

func TestEncodeWords_nil(t *testing.T) {
	if v := encodeWords(nil); v != nil {
		t.Fatal(v)
	}
}

func TestEncodeWords_siblings(t *testing.T) {
	node, err := Parse(strings.NewReader(`<div>one</div><div>two</div><div><div><div></div></div></div><div></div><div><div></div><div>three</div></div><div>four</div>`))
	if err != nil {
		t.Fatal(err)
	}
	if v := string(encodeWords(node.Data)); v != `one two three four` {
		t.Error(v)
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
	).OuterHTML(); v != `<b>a</b>` {
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
	if v := node.InnerWords(); v != `ONE TWO THREE FOUR !` {
		t.Fatal(v)
	}
	if v := node.InnerWords(func(node Node) bool {
		return node.Offset() == 0 &&
			node.Type() == html.TextNode
	}); v != `FOUR !` {
		t.Fatal(v)
	}
	if v := node.InnerWords(func(node Node) bool {
		return node.Offset() == 100
	}); v != `` {
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
				return node.Type() == html.ElementNode && node.Offset() == 2
			},
		)
	if len(nodes) != 2 {
		t.Fatal(len(nodes))
	}
	if v := nodes[0].Depth; v != 5 {
		t.Error(v)
	}
	if v := nodes[0].Offset(); v != 2 {
		t.Error(v)
	}
	if v := nodes[0].Match; v == nil {
		t.Error(v)
	} else {
		if v := v.Depth; v != 3 {
			t.Error(v)
		}
		if v := v.Offset(); v != 3 {
			t.Error(v)
		}
		if v := v.Match; v == nil || v.Data == nil {
			t.Error("expected initial matches to always be set for nodes with data")
		} else if v.Data.Type != html.DocumentNode {
			t.Fatal(v.Data.Type)
		} else if v.Match != nil {
			t.Error("expected the final match to be without unnecessary dupes")
		}
		if a := nodes[0].Parent().Parent(); a.Data == nil || a.Match != v || a.Match == v.Match {
			t.Error(a.Data)
		} else {
			if v := a.LastChild(); v.Depth != 4 {
				t.Error(v)
			} else if v.Match != a.Match {
				t.Error(v)
			} else if v := v.OuterHTML(); v != `<div a="5">five<div a="6">six</div></div>` {
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
		if v := v.OuterHTML(); v != `<div a="2">two</div>` {
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
	if v := nodes[0].OuterHTML(); v != `<div a="4">four</div>` {
		t.Error(v)
	}
	if v := nodes[1].OuterHTML(); v != `<div a="6">six</div>` {
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
	n, ok := parse(`<a><b></b><b><c id="search"></c></b><d><e></e><f><g></g></f></d></a>`).
		FindNode(
			func(node Node) bool {
				return node.GetAttrVal(``, `id`) == "search"
			},
		)
	if !ok || n.String() != `<c id="search"></c>` {
		t.Fatal(n, ok)
	}

	// parent filter test because lazy
	if n.Depth != 5 {
		t.Error(n.Depth)
	} else {
		if v := n.Parent(); v.Data == nil || v.Depth != 4 || v.OuterHTML() != `<b><c id="search"></c></b>` {
			t.Error(v.Data, v.Depth, v.OuterHTML())
		}
		if v := n.Parent(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil); v.Data == nil || v.Depth != 4 || v.OuterHTML() != `<b><c id="search"></c></b>` {
			t.Error(v.Data, v.Depth, v.OuterHTML())
		}
		if v := n.Parent(
			func(node Node) bool {
				return true
			},
		); v.Data == nil || v.Depth != 4 || v.OuterHTML() != `<b><c id="search"></c></b>` {
			t.Error(v.Data, v.Depth, v.OuterHTML())
		}
		if v := n.Parent(
			nil,
			nil,
			nil,
			func(node Node) bool {
				return node.Tag() == `b`
			},
		); v.Data == nil || v.Depth != 4 || v.OuterHTML() != `<b><c id="search"></c></b>` {
			t.Error(v.Data, v.Depth, v.OuterHTML())
		}
		if v := n.Parent(
			nil,
			nil,
			nil,
			func(node Node) bool {
				return node.Tag() == `body`
			},
		); v.Data == nil || v.Depth != 2 || v.OuterHTML() != `<body><a><b></b><b><c id="search"></c></b><d><e></e><f><g></g></f></d></a></body>` {
			t.Error(v.Data, v.Depth, v.OuterHTML())
		}
	}

	// sibling index and sibling length tests because lazy
	if v := n.SiblingIndex(); v != 0 {
		t.Error(v)
	}
	if v := n.SiblingLength(); v != 1 {
		t.Error(v)
	}
	if n := n.Parent(); n.Data == nil {
		t.Fatal(n)
	} else {
		// unfiltered values
		if v := n.SiblingIndex(); v != 1 {
			t.Error(v)
		}
		if v := n.SiblingLength(); v != 3 {
			t.Error(v)
		}

		// filtered values - use find on each node
		// - match on the current case (nested)
		if v := n.SiblingIndex(
			func(node Node) bool {
				return node.GetAttrVal(``, `id`) == "search"
			},
		); v != 0 {
			t.Error(v)
		}
		if v := n.SiblingLength(
			func(node Node) bool {
				return node.GetAttrVal(``, `id`) == "search"
			},
		); v != 1 {
			t.Error(v)
		}

		// - match on this and the previous sibling
		if v := n.SiblingIndex(
			func(node Node) bool {
				return node.Tag() == `b`
			},
		); v != 1 {
			t.Error(v)
		}
		if v := n.SiblingLength(
			func(node Node) bool {
				return node.Tag() == `b`
			},
		); v != 2 {
			t.Error(v)
		}

		// - match on the last sibling, but includes this node as well
		if v := n.SiblingIndex(
			func(node Node) bool {
				return node.Tag() == `d`
			},
		); v != 0 {
			t.Error(v)
		}
		if v := n.SiblingLength(
			func(node Node) bool {
				return node.Tag() == `d`
			},
		); v != 2 {
			t.Error(v)
		}

		// - no match, but the current node gets included in the length
		if v := n.SiblingIndex(
			func(node Node) bool {
				return false
			},
		); v != 0 {
			t.Error(v)
		}
		if v := n.SiblingLength(
			func(node Node) bool {
				return false
			},
		); v != 1 {
			t.Error(v)
		}

		// - nested match
		if v := n.SiblingIndex(
			func(node Node) bool {
				return node.Tag() == `b`
			},
			func(node Node) bool {
				return node.Tag() == `c`
			},
		); v != 0 {
			t.Error(v)
		}
		if v := n.SiblingLength(
			func(node Node) bool {
				return node.Tag() == `b`
			},
			func(node Node) bool {
				return node.Tag() == `c`
			},
		); v != 1 {
			t.Error(v)
		}

		if n := n.PrevSibling(); n.Data == nil {
			t.Fatal(n)
		} else {
			if v := n.SiblingIndex(); v != 0 {
				t.Error(v)
			}
			if v := n.SiblingLength(); v != 3 {
				t.Error(v)
			}
			if n := n.PrevSibling(); n.Data != nil {
				t.Fatal(n)
			} else {
				if v := n.SiblingIndex(); v != 0 {
					t.Error(v)
				}
				if v := n.SiblingLength(); v != 0 {
					t.Error(v)
				}
			}
		}
		if n := n.NextSibling(); n.Data == nil {
			t.Fatal(n)
		} else {
			// unfiltered values
			if v := n.SiblingIndex(); v != 2 {
				t.Error(v)
			}
			if v := n.SiblingLength(); v != 3 {
				t.Error(v)
			}

			// filtered values - use find on each node
			// - match + current node
			if v := n.SiblingIndex(
				func(node Node) bool {
					return node.GetAttrVal(``, `id`) == "search"
				},
			); v != 1 {
				t.Error(v)
			}
			if v := n.SiblingLength(
				func(node Node) bool {
					return node.GetAttrVal(``, `id`) == "search"
				},
			); v != 2 {
				t.Error(v)
			}

			// - again miss-match example
			if v := n.SiblingIndex(
				func(node Node) bool {
					return node.Tag() == `b`
				},
			); v != 2 {
				t.Error(v)
			}
			if v := n.SiblingLength(
				func(node Node) bool {
					return node.Tag() == `b`
				},
			); v != 3 {
				t.Error(v)
			}

			// - nested mis-match example
			if v := n.SiblingIndex(
				func(node Node) bool {
					return node.Tag() == `b`
				},
				func(node Node) bool {
					return node.Tag() == `c`
				},
			); v != 1 {
				t.Error(v)
			}
			if v := n.SiblingLength(
				func(node Node) bool {
					return node.Tag() == `b`
				},
				func(node Node) bool {
					return node.Tag() == `c`
				},
			); v != 2 {
				t.Error(v)
			}

			if n := n.FirstChild(); n.Data == nil {
				t.Fatal(n)
			} else {
				if v := n.SiblingIndex(); v != 0 {
					t.Error(v)
				}
				if v := n.SiblingLength(); v != 2 {
					t.Error(v)
				}
			}
			if n := n.NextSibling(); n.Data != nil {
				t.Fatal(n)
			} else {
				if v := n.SiblingIndex(); v != 0 {
					t.Error(v)
				}
				if v := n.SiblingLength(); v != 0 {
					t.Error(v)
				}
			}
		}
		if n := n.Parent(); n.Data == nil {
			t.Fatal(n)
		} else {
			if v := n.SiblingIndex(); v != 0 {
				t.Error(v)
			}
			if v := n.SiblingLength(); v != 1 {
				t.Error(v)
			}
		}
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
	if v := nodes[0].OuterHTML(); v != `<b></b>` {
		t.Fatal(v)
	}
	if v := nodes[1].Type(); v != html.TextNode {
		t.Fatal(v)
	}
	if v := nodes[1].OuterHTML(); v != ` ` {
		t.Fatal(v)
	}
	if v := nodes[2].Type(); v != html.ElementNode {
		t.Fatal(v)
	}
	if v := nodes[2].OuterHTML(); v != `<b><c id="search"></c></b>` {
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
		).OuterHTML(); v != `<b>a</b>` {
		t.Error(v)
	}
}

func TestNode_FindNode_nilNoMatchSet(t *testing.T) {
	n, ok := (Node{}).FindNode()
	if ok || n.Match != nil {
		t.Fatal(n, ok)
	}
}

func TestNode_Range_panic(t *testing.T) {
	defer func() {
		if r := fmt.Sprint(recover()); r != "htmlutil.Node.Range nil fn" {
			t.Fatal(r)
		}
	}()
	var n Node
	n.Range(nil)
}

func TestNode_Range_bailOutTrue(t *testing.T) {
	var (
		node   = parseElement(`<div><a>1</a><b>2</b><c>3</c></div>`)
		index  int
		values []string
	)
	node.Range(
		func(i int, node Node) bool {
			values = append(values, node.OuterHTML())
			if i != index {
				t.Fatal(i, index)
			}
			index++
			return index != 2
		},
	)
	if diff := deep.Equal(
		index,
		2,
	); diff != nil {
		t.Error(strings.Join(append([]string{"index diff:"}, diff...), "    \n"))
	}
	if diff := deep.Equal(
		values,
		[]string{
			`<a>1</a>`,
			`<b>2</b>`,
		},
	); diff != nil {
		t.Error(strings.Join(append([]string{"values diff:"}, diff...), "    \n"))
	}
}

func TestNode_Range_bailOutFalse(t *testing.T) {
	var (
		node   = parseElement(`<div><a>1</a><b>2</b><c>3</c></div>`)
		index  int
		values []string
	)
	node.Range(
		func(i int, node Node) bool {
			values = append(values, node.OuterHTML())
			if i != index {
				t.Fatal(i, index)
			}
			index++
			return true
		},
	)
	if diff := deep.Equal(
		index,
		3,
	); diff != nil {
		t.Error(strings.Join(append([]string{"index diff:"}, diff...), "    \n"))
	}
	if diff := deep.Equal(
		values,
		[]string{
			`<a>1</a>`,
			`<b>2</b>`,
			`<c>3</c>`,
		},
	); diff != nil {
		t.Error(strings.Join(append([]string{"values diff:"}, diff...), "    \n"))
	}
}

func TestNode_Parent_none(t *testing.T) {
	node := parse(`<a></a>`)
	if node.Depth != 0 {
		t.Error(node.Depth)
	}
	if node := node.Parent(); node.Data != nil || node.Depth != -1 {
		t.Error(node.Data, node.Depth)
	}
}

func TestNode_FirstChild_none(t *testing.T) {
	node := parseElement(`<a></a>`)
	if node.Depth != 3 {
		t.Error(node.Depth)
	}
	if node := node.FirstChild(); node.Data != nil || node.Depth != 4 {
		t.Error(node.Data, node.Depth)
	}
}

func TestNode_LastChild_none(t *testing.T) {
	node := parseElement(`<a></a>`)
	if node.Depth != 3 {
		t.Error(node.Depth)
	}
	if node := node.LastChild(); node.Data != nil || node.Depth != 4 {
		t.Error(node.Data, node.Depth)
	}
}

func TestNode_NextSibling_none(t *testing.T) {
	node := parseElement(`<a></a>`)
	if node.Depth != 3 {
		t.Error(node.Depth)
	}
	if node := node.NextSibling(); node.Data != nil || node.Depth != 3 {
		t.Error(node.Data, node.Depth)
	}
}

func TestNode_PrevSibling_none(t *testing.T) {
	node := parseElement(`<a></a>`)
	if node.Depth != 3 {
		t.Error(node.Depth)
	}
	if node := node.PrevSibling(); node.Data != nil || node.Depth != 3 {
		t.Error(node.Data, node.Depth)
	}
}

func TestNode_Range_filter(t *testing.T) {
	var (
		node   = parseElement(` <div>  <no>0</no><no>0</no> <a>1</a>   <b>2</b> <no>0</no><no>0</no>   <c>3</c>  <no>0</no><no>0</no>  </div>    `)
		index  int
		values []string
		count  int
		filter = func(node Node) bool {
			count++
			return node.Offset() == 0 &&
				node.Type() == html.ElementNode &&
				node.Tag() != `no`
		}
	)
	node.Range(
		func(i int, node Node) bool {
			values = append(values, node.OuterHTML())
			if i != index {
				t.Fatal(i, index)
			}
			index++
			return true
		},
		filter,
	)
	if diff := deep.Equal(
		index,
		3,
	); diff != nil {
		t.Error(strings.Join(append([]string{"index diff:"}, diff...), "    \n"))
	}
	if diff := deep.Equal(
		values,
		[]string{
			`<a>1</a>`,
			`<b>2</b>`,
			`<c>3</c>`,
		},
	); diff != nil {
		t.Error(strings.Join(append([]string{"values diff:"}, diff...), "    \n"))
	}
	if diff := deep.Equal(
		count,
		22,
	); diff != nil {
		t.Error(strings.Join(append([]string{"count diff:"}, diff...), "    \n"))
	}

	// children filter
	func() {
		var children []string
		for _, node := range node.Children(filter) {
			children = append(children, node.OuterHTML())
		}
		if len(children) != 3 {
			t.Error(len(children))
		}
		if diff := deep.Equal(
			children,
			values,
		); diff != nil {
			t.Error(strings.Join(append([]string{"children diff:"}, diff...), "    \n"))
		}
	}()

	// inner html filter
	func() {
		value := node.InnerHTML(filter)
		if diff := deep.Equal(
			value,
			strings.Join(values, ``),
		); diff != nil {
			t.Error(strings.Join(append([]string{"inner html diff:"}, diff...), "    \n"))
		}
	}()

	// inner text filter
	func() {
		value := node.InnerText(filter)
		if diff := deep.Equal(
			value,
			`123`,
		); diff != nil {
			t.Error(strings.Join(append([]string{"inner text diff:"}, diff...), "    \n"))
		}
	}()

	// iterate reverse filter
	func() {
		var children []string
		for node := node.LastChild(filter); node.Data != nil; node = node.PrevSibling(filter) {
			children = append([]string{node.OuterHTML()}, children...)
		}
		if len(children) != 3 {
			t.Error(len(children))
		}
		if diff := deep.Equal(
			children,
			values,
		); diff != nil {
			t.Error(strings.Join(append([]string{"children built from reverse diff:"}, diff...), "    \n"))
		}
	}()
}

func TestNode_HasClass(t *testing.T) {
	if v := parseElement(`<a class=""></a>`).HasClass(``); v {
		t.Error(v)
	}
	if v := parseElement(`<a class="  "></a>`).HasClass(`two`); v {
		t.Error(v)
	}
	if v := parseElement(`<a class="two"></a>`).HasClass(`two`); !v {
		t.Error(v)
	}
	if v := parseElement(`<a class="two"></a>`).HasClass(`TWO`); v {
		t.Error(v)
	}
	if v := parseElement(`<a class="one three"></a>`).HasClass(`two`); v {
		t.Error(v)
	}
	if v := parseElement(`<a class="one TWO three"></a>`).HasClass(`TWO`); !v {
		t.Error(v)
	}
	if v := parse(`<a class="two"></a>`).HasClass(`two`); v {
		t.Error(v)
	}
	if v := (Node{}).HasClass(`two`); v {
		t.Error(v)
	}
}

func TestNode_Classes(t *testing.T) {
	node := parseElement(`<a class="` + " \t\r\none \t\r\ntwo \t\r\n" + `"></a>`)
	if diff := deep.Equal(
		node.Classes(),
		[]string{
			`one`,
			`two`,
		},
	); diff != nil {
		t.Error(diff)
	}
}
