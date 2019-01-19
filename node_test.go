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
		func(node *html.Node) bool {
			return node.Type == html.ElementNode && node.Data == "div"
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

func TestNode_GetAttribute_caseInsensitive(t *testing.T) {
	node := parseElement(`<div one="value_1" TWO="VALUE_2"></div>`)
	if diff := deep.Equal(node.GetAttributeValue(``, `one`), `value_1`); diff != nil {
		t.Error(strings.Join(append([]string{"lowercase via lowercase diff:"}, diff...), "    \n"))
	}
	if diff := deep.Equal(node.GetAttributeValue(``, `ONE`), `value_1`); diff != nil {
		t.Error(strings.Join(append([]string{"lowercase via uppercase diff:"}, diff...), "    \n"))
	}
	if diff := deep.Equal(node.GetAttributeValue(``, `tWo`), `VALUE_2`); diff != nil {
		t.Error(strings.Join(append([]string{"uppercase via mixed diff:"}, diff...), "    \n"))
	}
}

func TestNode_GetAttribute_foreign(t *testing.T) {
	node := parseElement(`<svg viewBox="0 0 100 100" xlink:href="#icon-1" not="value"></svg>`)
	if diff := deep.Equal(node.GetAttributeValue(`xlink`, `href`), `#icon-1`); diff != nil {
		t.Error(strings.Join(append([]string{"match diff:"}, diff...), "    \n"))
	}
	if diff := deep.Equal(node.GetAttributeValue(`XLINK`, `href`), ``); diff != nil {
		t.Error(strings.Join(append([]string{"upper namespace diff:"}, diff...), "    \n"))
	}
	if diff := deep.Equal(node.GetAttributeValue(`xlink`, `HREF`), ``); diff != nil {
		t.Error(strings.Join(append([]string{"upper key diff:"}, diff...), "    \n"))
	}
	if diff := deep.Equal(node.GetAttributeValue(``, `href`), ``); diff != nil {
		t.Error(strings.Join(append([]string{"empty namespace diff:"}, diff...), "    \n"))
	}
}
