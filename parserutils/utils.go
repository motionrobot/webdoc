package parserutils

import (
	"fmt"
	"github.com/golang/glog"
	"golang.org/x/net/html"
	"strings"
)

func GetAttribute(n *html.Node, attr string) *html.Attribute {
	for _, a := range n.Attr {
		if a.Key == attr {
			return &a
		}
	}
	return nil
}

func AttributeValueMatch(n *html.Node, attr string, value string) *html.Attribute {
	for _, a := range n.Attr {
		if a.Key == attr {
			if a.Key == "class" {
				classes := strings.Split(a.Val, " ")
				for _, class := range classes {
					if len(class) > 0 && class == value {
						return &a
					}
				}
			} else if a.Val == value {
				return &a
			}
		}
	}
	return nil
}

func AttributeValueContains(n *html.Node, attr string, value string) *html.Attribute {
	for _, a := range n.Attr {
		if a.Key == attr {
			if strings.Contains(a.Val, value) {
				return &a
			}
		}
	}
	return nil
}

func IsAncestor(a *html.Node, d *html.Node) bool {
	n := d
	for n != nil && n.Type != html.DocumentNode {
		if n == a {
			return true
		}
		n = n.Parent
	}
	return false
}

func GetAncestors(n *html.Node) []*html.Node {
	results := make([]*html.Node, 0)
	for n != nil && n.Type != html.DocumentNode {
		results = append(results, n)
		n = n.Parent
	}
	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}
	glog.V(4).Info(GetDisplayNodePath(results))
	return results
}

func GetSharedAncestors(n1 *html.Node, n2 *html.Node) []*html.Node {
	a1, a2 := GetAncestors(n1), GetAncestors(n2)
	var i int
	for i = 0; i < len(a1) && i < len(a2); i++ {
		if a1[i] != a2[i] {
			break
		}
	}
	return a1[:i]
}

func GetRelatveNodePath(a *html.Node, d *html.Node) []*html.Node {
	ancestors := GetAncestors(d)
	for idx, n := range ancestors {
		if n == a {
			return ancestors[idx:]
		}
	}
	return nil
}

func GetDisplayAncestors(n *html.Node) string {
	ancestors := GetAncestors(n)
	return GetDisplayNodePath(ancestors)
}

func BuildDisplayNode(node *html.Node, b *strings.Builder) {
	fmt.Fprintf(b, "%v ", node.Type)
	switch node.Type {
	case html.ElementNode:
		fmt.Fprintf(b, "<%s>(%s)", node.DataAtom, node.Data)
	case html.TextNode:
		fmt.Fprintf(b, "<%s>(%s)", node.DataAtom, node.Data)
	default:
		fmt.Fprintf(b, "<%s>(%s)", node.DataAtom, node.Data)
	}
	class := GetAttribute(node, "class")
	if class != nil {
		fmt.Fprintf(b, "c:\"%s\" ", class.Val)
	}
	id := GetAttribute(node, "id")
	if id != nil {
		fmt.Fprintf(b, "id:\"%s\" ", id.Val)
	}
}

func GetDisplayNode(node *html.Node) string {
	var b strings.Builder
	b.Grow(32)
	BuildDisplayNode(node, &b)
	return b.String()
}

func GetDisplayNodePath(path []*html.Node) string {
	var b strings.Builder
	b.Grow(32)
	for _, node := range path {
		BuildDisplayNode(node, &b)
		fmt.Fprintf(&b, "->")
	}
	return b.String()
}

func BuildDisplayDescendants(node *html.Node, b *strings.Builder) {
	BuildDisplayDescendantsLevel(node, b, 0)
}

func BuildDisplayDescendantsLevel(node *html.Node, b *strings.Builder, level int) {
	fmt.Fprintf(b, strings.Repeat("++", level))
	BuildDisplayNode(node, b)
	fmt.Fprintf(b, "\n")
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		BuildDisplayDescendantsLevel(c, b, level+1)
	}
}

func GetDisplayDescendants(node *html.Node) string {
	var b strings.Builder
	b.Grow(32)
	BuildDisplayDescendants(node, &b)
	return b.String()
}