package extractor

import (
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
)

var (
	ImageCrawlInfoFilePtr = flag.String(
		"image_crawl_info_file",
		"",
		"The input file")
)

type ImageExtractor struct {
	cdoc *pb.CompositeDoc
	url  *url.URL
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
	ie.url = url

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

	if n.DataAtom.String() == "img" {
		glog.V(1).Infof("Found img tag")
		interested = true
	} else if n.DataAtom.String() == "image" {
		glog.V(1).Infof("Found image tag")
		utils.IncrementCounterNS("image", "all")
	} else if n.DataAtom.String() == "picture" {
		glog.V(1).Infof("Found picture tag")
		utils.IncrementCounterNS("picture", "all")
		interested = true
	}

	if interested {
		glog.V(1).Infof("%s===== %+v", displayPath, *n)
		glog.V(1).Info(pu.GetDisplayDescendants(n))
	} else {
		glog.V(2).Infof("%s===== %+v", displayPath, *n)
		glog.V(2).Info(pu.GetDisplayDescendants(n))
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

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ie.ProcessNode(c)
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
			srcUrl, urlErr = pu.GetAbsUrl(ie.url, src)
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
			dataSrcUrl, urlErr = pu.GetAbsUrl(ie.url, dataSrc)
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

	var imgSrcEles []*pb.ImageSrcEle
	if n.Parent != nil && n.Parent.DataAtom.String() == "picture" {
		imgSrcEles = ie.GetPictureSources(n.Parent)
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
		if len(imgSrcEles) > 0 {
			glog.V(1).Infof("picture-srcset-unresolved: %s", dataSrcSet)
		}
	}
}

func (ie *ImageExtractor) GetPictureSources(n *html.Node) []*pb.ImageSrcEle {
	imgSrcEles := make([]*pb.ImageSrcEle, 0)
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.DataAtom.String() == "source" {
			imgSrcEle := &pb.ImageSrcEle{}
			glog.V(1).Infof("Getting source node:\n%s",
				pu.GetDisplayAttributes(c))
			imgSrcEles = append(imgSrcEles, imgSrcEle)
			for _, a := range c.Attr {
				if a.Key == "srcset" {
					imgSrcEle.Url = a.Val
					eles, err := pu.ParseSrcSet(a.Val)
					if err != nil {
						glog.Fatal("%s", a.Val, err)
					}
					if len(eles) == 0 {
						glog.V(1).Info("Empty srcset")
					}
					for _, ele := range eles {
						glog.V(1).Infof("Getting src:\n%s",
							proto.MarshalTextString(ele))
					}
				} else if a.Key == "sizes" {
					imgSrcEle.SizeDesc = a.Val
				} else {
					utils.IncrementCounterNS(
						"source",
						fmt.Sprintf("attr_%s", a.Key))
				}
			}
			if len(imgSrcEle.GetUrl()) == 0 {
				utils.IncrementCounterNS("source", "srcset-missing")
			}
			if len(imgSrcEle.GetSizeDesc()) == 0 {
				utils.IncrementCounterNS("source", "media-missing")
			}
		}
	}
	return imgSrcEles
}
