package htmlutil

import (
	"golang.org/x/net/html"
)

func GetAttribute(namespace string, key string, attributes ...html.Attribute) (html.Attribute, bool) {
	for _, attribute := range attributes {
		if attribute.Namespace == namespace && attribute.Key == key {
			return attribute, true
		}
	}
	return html.Attribute{}, false
}

func GetAttributeValue(namespace string, key string, attributes ...html.Attribute) string {
	result, _ := GetAttribute(namespace, key, attributes...)
	return result.Val
}

func FilterElements(node *html.Node, filters ...func(node *html.Node) bool) []*html.Node {
	return filterElements(
		filterElementsConfig{
			Node:    node,
			Filters: filters,
		},
	)
}

func FindElement(node *html.Node, filters ...func(node *html.Node) bool) (*html.Node, bool) {
	elements := filterElements(
		filterElementsConfig{
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
