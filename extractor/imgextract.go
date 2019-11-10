package extractor

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/motionrobot/utils"
	pu "github.com/motionrobot/webdoc/parserutils"
	pb "github.com/motionrobot/webdoc/proto"
	"golang.org/x/net/html"
	"io"
	"net/url"
	"strings"
)

var (
	ImageCrawlInfoFilePtr = flag.String(
		"image_crawl_info_file",
		"",
		"The input file")
)

type ImageExtractor struct {
	cdoc   *pb.CompositeDoc
	docUrl *url.URL
}

func NewImageExtractor() *ImageExtractor {
	return &ImageExtractor{}
}

func (ie *ImageExtractor) Reset() {
	ie.cdoc = &pb.CompositeDoc{
		Images: make([]*pb.ImageElement, 0),
	}
}

func (ie *ImageExtractor) Finalize() {
	glog.V(0).Infof("File has %d images found", len(ie.cdoc.GetImages()))
	glog.V(0).Infof("Composite doc:\n%s", proto.MarshalTextString(ie.cdoc))
}

func (ie *ImageExtractor) Parse(r io.Reader, cdoc *pb.CompositeDoc) error {
	ie.cdoc = cdoc
	if ie.cdoc == nil {
		ie.cdoc = &pb.CompositeDoc{Images: make([]*pb.ImageElement, 0)}
	}
	url, err := url.Parse(cdoc.GetUrl())
	if err != nil {
		glog.Fatal(err)
	}
	glog.V(1).Infof("Extracting images from %s", url.String())
	ie.docUrl = url

	utils.IncrementCounterNS("doc", "ParsedDoc")

	doc, err := html.Parse(r)
	if err != nil {
		glog.Fatal(err)
	}
	ie.ProcessNode(doc)
	return nil
}

func (ie *ImageExtractor) GetDoc() *pb.CompositeDoc {
	return ie.cdoc
}

func (ie *ImageExtractor) ProcessNode(n *html.Node) {
	displayPath := pu.GetDisplayAncestors(n)
	interested := false
	var err error
	var noscriptNode *html.Node

	switch n.DataAtom.String() {
	case "img":
		glog.V(1).Infof("Found img tag")
		interested = true
	case "image":
		glog.V(1).Infof("Found image tag")
		utils.IncrementCounterNS("image", "all")
	case "picture":
		glog.V(1).Infof("Found picture tag")
		utils.IncrementCounterNS("picture", "all")
		interested = true
	case "noscript":
		noscriptNode, err = ie.ProcessNoscriptNode(n)
		if err != nil {
			glog.Fatal(err)
		}
	case "script":
		typeStr, err := pu.GetAttributeValue(n, "type")
		if err == nil && len(typeStr) > 0 {
			utils.IncrementCounterNS("script", typeStr)
			if typeStr == "application/ld+json" {
				ie.ProcessScriptNode(n)
			}
		}
	}

	if interested {
		glog.V(1).Infof("%s===== %+v", displayPath, *n)
		if n.FirstChild != nil {
			glog.V(1).Info(pu.GetDisplayDescendants(n, 2, false))
		}
	} else {
		glog.V(2).Infof("%s===== %+v", displayPath, *n)
		glog.V(2).Info(pu.GetDisplayDescendants(n, 2, false))
	}

	if n.DataAtom.String() == "img" {
		utils.IncrementCounterNS("img", "all")
		imgEle := &pb.ImageElement{}
		ie.cdoc.Images = append(ie.cdoc.Images, imgEle)

		height, err := pu.GetAttributeIntValue(n, "height")
		switch err {
		case pu.ErrAttrNotFound:
			utils.IncrementCounterNS("img", "height-no")
			glog.V(1).Infof("Image Element has no height")
		case pu.ErrAttrMalFormatted:
			utils.IncrementCounterNS("img", "height-bad")
			glog.V(1).Infof("Image Element has bad height")
		case nil:
			imgEle.Height = int32(height)
		default:
			glog.Fatal(err)
		}

		width, err := pu.GetAttributeIntValue(n, "width")
		switch err {
		case pu.ErrAttrNotFound:
			utils.IncrementCounterNS("img", "width-no")
			glog.V(1).Infof("Image Element has no width")
		case pu.ErrAttrMalFormatted:
			utils.IncrementCounterNS("img", "width-bad")
			glog.V(1).Infof("Image Element has bad width")
		case nil:
			imgEle.Width = int32(width)
		default:
			glog.Fatal(err)
		}

		alt, err := pu.GetAttributeValue(n, "alt")
		switch err {
		case pu.ErrAttrNotFound:
			glog.V(1).Infof("Image Element has no alt")
		case nil:
			imgEle.Alt = alt
		default:
			glog.Fatal(err)
		}
		ie.FilImageUrl(n, imgEle)

		glog.V(1).Infof("Getting image element:\n%s",
			proto.MarshalTextString(imgEle))
	}

	if noscriptNode != nil {
		ie.ProcessNode(noscriptNode)
	} else {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			ie.ProcessNode(c)
		}
	}
}

func (ie *ImageExtractor) FilImageUrl(n *html.Node, imgEle *pb.ImageElement) {
	var url, srcUrl, dataSrcUrl *url.URL
	var urlErr error

	src, err := pu.GetAttributeValue(n, "src")
	switch err {
	case pu.ErrAttrNotFound:
		glog.V(1).Infof("Image Element has no src")
	case nil:
		if len(src) != 0 {
			srcUrl, urlErr = pu.GetAbsUrl(ie.docUrl, src)
			if urlErr != nil {
				glog.Fatal(urlErr)
			}
			utils.IncrementCounterNS("img", "src-good")
		} else {
			utils.IncrementCounterNS("img", "src-no")
			glog.V(1).Infof("Image Element has empty src")
		}
	default:
		glog.Fatal(err)
	}

	dataSrc, err := pu.GetAttributeValue(n, "data-src")
	switch err {
	case nil:
		if len(dataSrc) != 0 {
			dataSrcUrl, urlErr = pu.GetAbsUrl(ie.docUrl, dataSrc)
			if urlErr != nil {
				glog.Fatal(urlErr)
			}
			utils.IncrementCounterNS("img", "data-src-good")
		} else {
			glog.V(1).Infof("Image Element has empty data-src")
		}
	case pu.ErrAttrNotFound:
	default:
		glog.Fatal(err)
	}

	if srcUrl != nil && dataSrcUrl != nil && srcUrl.String() != dataSrcUrl.String() {
		glog.V(1).Infof("Trying to pick from src %s and data-src %s",
			srcUrl.String(), dataSrcUrl.String())
		glog.V(1).Infof("Scheme %s vs. %s", srcUrl.Scheme, dataSrcUrl.Scheme)
		if srcUrl.Scheme == "data" && dataSrcUrl.Scheme != "data" {
			url = dataSrcUrl
		} else {
			glog.V(1).Infof("Unresolved src %s and data-src %s",
				srcUrl.String(), dataSrcUrl.String())
			utils.IncrementCounterNS("img", "src-unresolved")
		}
	} else if srcUrl != nil {
		url = srcUrl
	} else if dataSrcUrl != nil {
		url = dataSrcUrl
	}

	if url != nil {
		imgEle.Url = url.String()
		utils.IncrementCounterNS("img", fmt.Sprintf("scheme_%s", url.Scheme))
		if url.Scheme == "data" {
			glog.V(1).Infof("Data scheme for src %s", url.String())
		}
	} else {
		utils.IncrementCounterNS("img", "url-not-set")
	}

	imgGroupInfos := make([]*pb.ImageGroupInfo, 0)
	if n.Parent != nil && n.Parent.DataAtom.String() == "picture" {
		imgGroupInfos = ie.GetPictureSources(n.Parent)
	}

	srcSet, err := pu.GetAttributeValue(n, "srcset")
	switch err {
	case nil:
		glog.V(1).Infof("srcset: %s", srcSet)
		utils.IncrementCounterNS("img", "srcset")
	case pu.ErrAttrNotFound:
	default:
		glog.Fatal(err)

	}

	dataSrcSet, err := pu.GetAttributeValue(n, "data-srcset")
	switch err {
	case nil:
		glog.V(1).Infof("data-srcset: %s", dataSrcSet)
		utils.IncrementCounterNS("img", "data-srcset")
	case pu.ErrAttrNotFound:
	default:
		glog.Fatal(err)
	}

	var srcSetFinal string
	if len(srcSet) > 0 && len(dataSrcSet) > 0 {
		glog.V(1).Infof("srcset-unresolved: %s", dataSrcSet)
	} else {
		if len(srcSet) > 0 {
			srcSetFinal = srcSet
		} else if len(dataSrcSet) > 0 {
			srcSetFinal = dataSrcSet
		}
	}
	if len(srcSetFinal) > 0 {
		if len(imgGroupInfos) > 0 {
			glog.V(1).Infof("picture-srcset-unresolved: %s", srcSetFinal)
		} else {
			imgGroupInfo := &pb.ImageGroupInfo{}
			imgGroupInfo.ImageSources = ie.ParseSrcSet(srcSetFinal)
			imgGroupInfos = append(imgGroupInfos, imgGroupInfo)
		}
	}
	if len(imgGroupInfos) > 0 {
		imgEle.ImageGroups = imgGroupInfos
	}
}

func (ie *ImageExtractor) ProcessScriptNode(n *html.Node) error {
	if n.DataAtom.String() != "script" {
		glog.Fatal("Shouldn't be here, this is just for script")
	}
	if n.FirstChild == nil {
		return nil
	}
	if n.FirstChild != n.LastChild {
		utils.IncrementCounterNS("script", "multiple-child")
		return nil
	}
	if n.FirstChild.Type != html.TextNode {
		utils.IncrementCounterNS("script", "non-text-child")
		return nil
	}
	var v interface{}
	json.Unmarshal([]byte(n.FirstChild.Data), &v)
	data := v.(map[string]interface{})
	for k, v := range data {
		utils.IncrementCounterNS("script", k)
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
	return nil
}

func (ie *ImageExtractor) ProcessNoscriptNode(n *html.Node) (*html.Node, error) {
	if n.DataAtom.String() != "noscript" {
		glog.Fatal("Shouldn't be here, this is just for noscript")
	}
	if n.FirstChild == nil {
		return nil, nil
	}
	if n.FirstChild != n.LastChild {
		utils.IncrementCounterNS("noscript", "multiple-child")
		return nil, nil
	}
	if n.FirstChild.Type != html.TextNode {
		utils.IncrementCounterNS("noscript", "non-text-child")
		return nil, nil
	}
	r := strings.NewReader(n.FirstChild.Data)
	doc, err := html.Parse(r)
	if err != nil {
		utils.IncrementCounterNS("noscript", "expanded-failed")
	} else {
		utils.IncrementCounterNS("noscript", "expanded")
	}
	return doc, err
}

func (ie *ImageExtractor) GetPictureSources(n *html.Node) []*pb.ImageGroupInfo {
	if n.DataAtom.String() != "picture" {
		glog.Fatal("Shouldn't be here, this is just for picture")
	}
	glog.V(1).Infof("Picture tag:\n%s", pu.GetLongDisplayNode(n))
	glog.V(1).Infof("Picture tag children:\n%s", pu.GetDisplayDescendants(n, 2, true))
	imgGroupInfos := make([]*pb.ImageGroupInfo, 0)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		glog.V(1).Infof("Picture child tag:\n%s", pu.GetLongDisplayNode(c))
		if c.DataAtom.String() == "source" {
			imgGroupInfo := ie.GetImageGroupFromSource(c)
			imgGroupInfos = append(imgGroupInfos, imgGroupInfo)
		}
	}
	if len(imgGroupInfos) == 0 {
		glog.V(1).Infof("Empty source with no urls")
	}
	return imgGroupInfos
}

func (ie *ImageExtractor) GetImageGroupFromSource(n *html.Node) *pb.ImageGroupInfo {
	imgGroupInfo := &pb.ImageGroupInfo{}
	glog.V(1).Infof("Getting source node:\n%s", pu.GetDisplayAttributes(n))
	srcSet := ""
	dataSrcSet := ""
	for _, a := range n.Attr {
		if a.Key == "type" {
			imgGroupInfo.Type = a.Val
		} else if a.Key == "media" {
			imgGroupInfo.Media = a.Val
		} else if a.Key == "srcset" {
			srcSet = a.Val
		} else if a.Key == "data-srcset" {
			dataSrcSet = a.Val
		}
		utils.IncrementCounterNS(
			"source",
			fmt.Sprintf("attr_%s", a.Key))
	}
	var srcSetEles, dataSrcSetEles []*pb.ImageSrcEle
	if len(srcSet) > 0 {
		srcSetEles = ie.ParseSrcSet(srcSet)
	}
	if len(dataSrcSet) > 0 {
		dataSrcSetEles = ie.ParseSrcSet(dataSrcSet)
	}
	if len(srcSetEles) > 0 && len(dataSrcSetEles) > 0 {
		utils.IncrementCounterNS(
			"source",
			"srcset-unresolved")
	} else {
		if len(srcSetEles) > 0 {
			imgGroupInfo.ImageSources = srcSetEles
		} else if len(dataSrcSetEles) > 0 {
			imgGroupInfo.ImageSources = dataSrcSetEles
		}
	}
	glog.V(1).Infof("Getting image group info from source:\n%s",
		proto.MarshalTextString(imgGroupInfo))
	return imgGroupInfo
}

func (ie *ImageExtractor) ParseSrcSet(srcset string) []*pb.ImageSrcEle {
	imgSrcEles, err := pu.ParseSrcSet(srcset)
	if err != nil {
		glog.Fatal(err)
	}
	for _, imgSrcEle := range imgSrcEles {
		imgUrl, urlErr := pu.GetAbsUrl(ie.docUrl, imgSrcEle.GetUrl())
		if urlErr != nil {
			glog.Fatal(urlErr)
		}
		imgSrcEle.Url = imgUrl.String()
	}
	return imgSrcEles
}
