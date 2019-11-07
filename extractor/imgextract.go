package extractor

import (
	"flag"
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
		interested = true
	} else if n.DataAtom.String() == "image" {
		glog.V(1).Infof("Found image tags")
		utils.IncrementCounterNS("image", "all")
	}

	if interested {
		glog.V(1).Infof("%s===== %+v", displayPath, *n)
	} else {
		glog.V(2).Infof("%s===== %+v", displayPath, *n)
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
		glog.V(1).Infof("Getting image element:\n%s",
			proto.MarshalTextString(imgEle))

		ie.FilImageUrl(n, imgEle)
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

	if srcUrl != nil && dataSrcUrl != nil && srcUrl.String() == dataSrcUrl.String() {
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
	} else if dataSrcUrl != nil {
		url = dataSrcUrl
	} else if srcUrl != nil {
		url = srcUrl
	}

	if url != nil {
		imgEle.Url = url.String()
		utils.IncrementCounterNS("scheme", url.Scheme)
		if url.Scheme == "data" {
			glog.V(1).Infof("Data scheme for src %s", url.String())
		}
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

}
