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
	"github.com/go-test/deep"
	"golang.org/x/net/html"
	"strings"
	"testing"
)

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
