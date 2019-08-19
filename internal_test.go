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
	"golang.org/x/net/html"
	"testing"
)

func filterNodesRaw(node *html.Node, filters ...func(node Node) bool) []Node {
	return filterNodes(Node{Data: node}, filters...)
}

func findNodeRaw(node *html.Node, filters ...func(node Node) bool) (Node, bool) {
	return findNode(Node{Data: node}, filters...)
}

func getNodeRaw(node *html.Node, filters ...func(node Node) bool) Node {
	return getNode(Node{Data: node}, filters...)
}

func TestSiblingIndex_nil(t *testing.T) {
	if v := siblingIndex(Node{}); v != 0 {
		t.Fatal(v)
	}
}

func TestSiblingLength_nil(t *testing.T) {
	if v := siblingLength(Node{}); v != 0 {
		t.Fatal(v)
	}
}
