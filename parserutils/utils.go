package parserutils

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/golang/glog"
	pb "github.com/motionrobot/webdoc/proto"
	"golang.org/x/net/html"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var ErrAttrNotFound = errors.New("not found")
var ErrAttrMalFormatted = errors.New("mal formatted")

type Parser interface {
	Reset()
	Parse(io.Reader, *pb.CompositeDoc) error
	Finalize()
}

func ParseFile(fn string, p Parser, doc *pb.CompositeDoc) error {
	f, err := os.Open(fn)
	if err != nil {
		glog.Fatal(err)
	}
	reader := bufio.NewReader(f)
	return p.Parse(reader, doc)
}

// Attributes
func GetAttribute(n *html.Node, attr string) *html.Attribute {
	for _, a := range n.Attr {
		if a.Key == attr {
			return &a
		}
	}
	return nil
}

func GetAttributeValue(n *html.Node, attr string) (string, error) {
	a := GetAttribute(n, attr)
	if a != nil {
		return strings.TrimSpace(a.Val), nil
	}
	return "", ErrAttrNotFound
}

func GetAttributeIntValue(n *html.Node, attr string) (int64, error) {
	val, err := GetAttributeValue(n, attr)
	if err != nil {
		return 0, err
	}
	intVal, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, ErrAttrMalFormatted
	}
	return intVal, nil
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

func BuildDisplayAttribute(attr html.Attribute, b *strings.Builder, level int) {
	fmt.Fprintf(b, strings.Repeat("  ", level))
	fmt.Fprintf(b, "%s:%s %s\n", attr.Namespace, attr.Key, attr.Val)
}

func GetDisplayAttribute(attr html.Attribute, level int) string {
	var b strings.Builder
	b.Grow(32)
	BuildDisplayAttribute(attr, &b, level)
	return b.String()
}

func BuildDisplayAttributes(node *html.Node, b *strings.Builder, level int) {
	for _, attr := range node.Attr {
		BuildDisplayAttribute(attr, b, level)
	}
}

func GetDisplayAttributes(node *html.Node) string {
	var b strings.Builder
	b.Grow(32)
	BuildDisplayAttributes(node, &b, 0)
	return b.String()
}

func ParseSrcSet(srcset string) ([]*pb.ImageSrcEle, error) {
	imgSrcEles := make([]*pb.ImageSrcEle, 0)
	segments := strings.Split(strings.TrimSpace(srcset), ",")
	for _, segment := range segments {
		segs := strings.Split(strings.TrimSpace(segment), " ")
		if len(segs) > 2 {
			glog.V(0).Infof("mal-formatted segments: %s", segment)
			return nil, ErrAttrMalFormatted
		}
		imgSrcEle := &pb.ImageSrcEle{}
		imgSrcEle.Url = segs[0]
		if len(segs) == 2 {
			imgSrcEle.SizeDesc = segs[1]
		}
		imgSrcEles = append(imgSrcEles, imgSrcEle)
	}
	return imgSrcEles, nil
}

func FixUrl(rawurl string) string {
	return strings.TrimSpace(rawurl)
}

func GetAbsUrl(baseUrl *url.URL, ref string) (*url.URL, error) {
	var resultUrl *url.URL
	var err error
	fixedRef := FixUrl(ref)
	glog.V(1).Infof("Maybe fix relative url %s with %s",
		fixedRef, baseUrl.String())
	resultUrl, err = url.Parse(fixedRef)
	if err != nil {
		return nil, err
	}
	if !resultUrl.IsAbs() {
		glog.V(1).Infof("Found relative url %s", ref)
		resultUrl, err = baseUrl.Parse(resultUrl.String())
		if err != nil {
			return nil, err
		}
		glog.V(1).Infof("Absolute URL is %s", resultUrl.String())
	}
	return resultUrl, nil
}

// Nodes
func BuildDisplayNode(node *html.Node, b *strings.Builder, level int, long bool) {
	fmt.Fprintf(b, "%v ", node.Type)
	switch node.Type {
	case html.ElementNode:
		fmt.Fprintf(b, "<%s>(%s)", node.DataAtom, node.Data)
	case html.TextNode:
		fmt.Fprintf(b, "<%s>(%s)", node.DataAtom, node.Data)
	default:
		fmt.Fprintf(b, "<%s>(%s)", node.DataAtom, node.Data)
	}
	if long {
		fmt.Fprint(b, "\n")
		BuildDisplayAttributes(node, b, 1)
	} else {
		class := GetAttribute(node, "class")
		if class != nil {
			fmt.Fprintf(b, "c:\"%s\" ", class.Val)
		}
		id := GetAttribute(node, "id")
		if id != nil {
			fmt.Fprintf(b, "id:\"%s\" ", id.Val)
		}
	}
}

func GetDisplayNode(node *html.Node) string {
	var b strings.Builder
	b.Grow(32)
	BuildDisplayNode(node, &b, 0, false)
	return b.String()
}

func GetLongDisplayNode(node *html.Node) string {
	var b strings.Builder
	b.Grow(32)
	BuildDisplayNode(node, &b, 0, true)
	return b.String()
}

// Ancestors and descendents
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

func GetDisplayNodePath(path []*html.Node) string {
	var b strings.Builder
	b.Grow(32)
	for _, node := range path {
		BuildDisplayNode(node, &b, 0, false)
		fmt.Fprintf(&b, "->")
	}
	return b.String()
}

func BuildDisplayDescendants(node *html.Node, b *strings.Builder, level int, long bool) {
	fmt.Fprintf(b, strings.Repeat("++", level))
	BuildDisplayNode(node, b, level, long)
	fmt.Fprintf(b, "\n")
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		BuildDisplayDescendants(c, b, level+1, long)
	}
}

func GetDisplayDescendants(node *html.Node, long bool) string {
	var b strings.Builder
	b.Grow(32)
	BuildDisplayDescendants(node, &b, 0, long)
	return b.String()
}
