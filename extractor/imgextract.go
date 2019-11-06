package extractor

import (
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	pu "github.com/motionrobot/webdoc/parserutils"
	pb "github.com/motionrobot/webdoc/proto"
	"golang.org/x/net/html"
	"io"
)

type ImageExtractor struct {
	doc *pb.CompositeDoc
}

func NewImageExtractor() *ImageExtractor {
	return &ImageExtractor{}
}

func (ie *ImageExtractor) Reset() {
	ie.doc = &pb.CompositeDoc{
		Images: make([]*pb.ImageElement, 0),
	}
}

func (ie *ImageExtractor) Finalize() {
	glog.V(0).Infof("File has %d images found", len(ie.doc.GetImages()))
	glog.V(0).Infof("Composite doc:\n%s", proto.MarshalTextString(ie.doc))
}

func (ie *ImageExtractor) Parse(r io.Reader) error {
	doc, err := html.Parse(r)
	if err != nil {
		glog.Fatal(err)
	}
	ie.ProcessNode(doc)
	return nil
}

func (ie *ImageExtractor) ProcessNode(n *html.Node) {
	displayPath := pu.GetDisplayAncestors(n)
	interested := false

	if n.DataAtom.String() == "img" {
		imgEle := &pb.ImageElement{}
		ie.doc.Images = append(ie.doc.Images, imgEle)

		height, err := pu.GetAttributeIntValue(n, "height")
		switch err {
		case pu.ErrAttrNotFound:
			glog.V(1).Infof("Image Element has no height")
		case pu.ErrAttrMalFormatted:
			glog.V(1).Infof("Image Element has no height")
		case nil:
			imgEle.Height = int32(height)
		default:
			glog.Fatal(err)
		}

		width, err := pu.GetAttributeIntValue(n, "width")
		switch err {
		case pu.ErrAttrNotFound:
			glog.V(1).Infof("Image Element has no width")
		case pu.ErrAttrMalFormatted:
			glog.V(1).Infof("Image Element has no width")
		case nil:
			imgEle.Width = int32(width)
		default:
			glog.Fatal(err)
		}

		src, err := pu.GetAttributeValue(n, "src")
		switch err {
		case pu.ErrAttrNotFound:
			glog.V(1).Infof("Image Element has no src")
		case nil:
			imgEle.Url = src
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

		interested = true
	}

	if interested {
		glog.V(1).Infof("%s===== %+v", displayPath, *n)
	} else {
		glog.V(2).Infof("%s===== %+v", displayPath, *n)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ie.ProcessNode(c)
	}
}
