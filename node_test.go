package htmlutil

import (
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
