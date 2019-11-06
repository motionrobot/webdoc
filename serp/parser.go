package serp

import (
	"bufio"
	"encoding/json"
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	pu "github.com/motionrobot/webdoc/parserutils"
	pb "github.com/motionrobot/webdoc/proto"
	"golang.org/x/net/html"
	"io"
	"os"
	"strconv"
)

type ResultJson struct {
	Isu          string `json:"isu"`
	Url          string `json:"ru"`
	Snippet      string `json:"pt"`
	Id           string `json:"id"`
	ImageUrl     string `json:"ou"`
	ThumbnailUrl string `json:"tu"`
	SiteName     string `json:"st"`
	Site         string `json:"rh"`
}

/*
type ResultJson struct {
	Isu          string `json:"isu"`
	Url          string `json:"ru"`
	Snippet      string `json:"pt"`
	Id           string `json:"id"`
	Cb  int
	Cl  int
	Clt string
	Cr  int
	Ct  int
	Dd  string
	Itg int
	Ity string
	Oh  int
	Ou  string
	Ow  int
	Rh  string
	Rid string
	Rmt int
	Rt  int
	Sc  int
	St  string
	Th  int
	Tu  string
	Tw  int
}
*/

type SERPParser struct {
	resultRoot    *html.Node
	resultNodes   []*html.Node
	curResultNode *html.Node
	resultPage    *pb.GenericSearchResponse
}

func NewSERPParser() *SERPParser {
	return &SERPParser{
		resultNodes: make([]*html.Node, 0),
		resultPage:  NewResultPage()}
}

func (p *SERPParser) Reset() {
	p.resultRoot = nil
	p.curResultNode = nil
	p.resultNodes = make([]*html.Node, 0)
	p.resultPage = NewResultPage()
}

func (p *SERPParser) ParseFile(fn string) error {
	f, err := os.Open(fn)
	if err != nil {
		glog.Fatal(err)
	}
	reader := bufio.NewReader(f)
	return p.Parse(reader)
}

func (p *SERPParser) Parse(r io.Reader) error {
	doc, err := html.Parse(r)
	if err != nil {
		glog.Fatal(err)
	}
	p.ProcessNode(doc)
	return nil
}

func (p *SERPParser) Finalize() {
	if p.resultPage != nil {
		glog.V(0).Infof("Result page:\n%s", proto.MarshalTextString(p.resultPage))
	}
}

func (p *SERPParser) GetResultPage() *pb.GenericSearchResponse {
	return p.resultPage
}

func (p *SERPParser) GetCurResult() *pb.Result {
	if len(p.resultPage.Results) == 0 {
		return nil
	}
	return p.resultPage.GetResults()[len(p.resultPage.Results)-1]
}

func (p *SERPParser) ProcessNode(n *html.Node) {
	displayPath := pu.GetDisplayAncestors(n)
	interested := false
	isResultNode := false

	if pu.AttributeValueMatch(n, "id", "search") != nil {
		if p.resultRoot != nil {
			glog.Fatalf("Two roots:\n%s\n%s",
				pu.GetDisplayAncestors(p.resultRoot),
				displayPath)
		}
		// It is the root of search result section
		p.resultRoot = n
		interested = true
	} else if attr := pu.GetAttribute(n, "data-ri"); attr != nil {
		if p.resultRoot != nil && pu.IsAncestor(p.resultRoot, n) {
			// It is the root of a search result
			p.resultNodes = append(p.resultNodes, n)
			p.curResultNode = n
			isResultNode = true
			interested = true
			// We find a new result
			pos, err := strconv.ParseUint(attr.Val, 10, 32)
			if err != nil {
				glog.Fatal(err)
			}
			result := &pb.Result{Pos: uint32(pos)}
			p.resultPage.Results = append(
				p.resultPage.Results, result)

			glog.V(1).Infof("Result node:\n%s", pu.GetDisplayDescendants(n))
			glog.V(1).Infof("Result node:\n%s", displayPath)
		}
	} else if p.curResultNode != nil {
		if n.DataAtom.String() == "div" && pu.AttributeValueMatch(n, "class", "notranslate") != nil {
			glog.V(0).Infof("%s", n.FirstChild.Data)
			resultJson, err := p.ParseResultJson(n.FirstChild)
			if err != nil {
				glog.Fatal(err)
			}
			glog.V(0).Infof("%+v", *resultJson)
			result := p.GetCurResult()
			FillResultWithJson(result, resultJson)
		}

		interested = true
	}
	if interested {
		glog.V(1).Infof("%s===== %+v", displayPath, *n)
	} else {
		glog.V(2).Infof("%s===== %+v", displayPath, *n)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.ProcessNode(c)
	}
	if isResultNode {
		p.curResultNode = nil
	}
}

func FillResultWithJson(result *pb.Result, resultJson *ResultJson) {
	result.Snippet = resultJson.Snippet
	result.Url = resultJson.Url
	result.ThumbnailUrl = resultJson.ThumbnailUrl
	result.ImageUrl = resultJson.ImageUrl
	result.Site = resultJson.Site
}

func (p *SERPParser) ParseResultJson(node *html.Node) (*ResultJson, error) {
	if node.Type != html.TextNode {
		return nil, nil
	}
	result := &ResultJson{}
	glog.V(0).Infof("%s: %v", node.Data, json.Valid([]byte(node.Data)))
	err := json.Unmarshal([]byte(node.Data), result)

	if glog.V(1) {
		var v interface{}
		json.Unmarshal([]byte(node.Data), &v)
		data := v.(map[string]interface{})
		for k, v := range data {
			switch v := v.(type) {
			case string:
				glog.V(0).Infof("%v %v (string)", k, v)
			case float64:
				glog.V(0).Infof("%v %v (float64)", k, v)
			case []interface{}:
				glog.V(0).Infof("%v (array):", k)
				for i, u := range v {
					glog.V(0).Info("    ", i, u)
				}
			default:
			}
		}
	}
	if result.Isu != result.Site {
		glog.V(0).Infof("Not equal: %s vs %s", result.Isu, result.Site)
	}
	return result, err
}

func NewResultPage() *pb.GenericSearchResponse {
	p := &pb.GenericSearchResponse{}
	p.Results = make([]*pb.Result, 0, 20)
	return p
}
