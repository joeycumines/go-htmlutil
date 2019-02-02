package htmlutil

import "testing"

func TestSiblingIndex_nil(t *testing.T) {
	if v := siblingIndex(nil); v != 0 {
		t.Fatal(v)
	}
}

func TestSiblingLength_nil(t *testing.T) {
	if v := siblingLength(nil); v != 0 {
		t.Fatal(v)
	}
}
