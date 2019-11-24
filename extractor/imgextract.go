package extractor

import (
	// "encoding/json"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/html"
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/motionrobot/utils"
	pu "github.com/motionrobot/webdoc/parserutils"
	pb "github.com/motionrobot/webdoc/proto"
)

var (
	ImageCrawlInfoFilePtr = flag.String(
		"image_crawl_info_file",
		"",
		"The input file")
)

type ImageFileType int32

const (
	ImageFileType_UNKNOWN ImageFileType = iota
	ImageFileType_JPG
	ImageFileType_PNG
	ImageFileType_GIF
)

type ImageExtractor struct {
	cdoc         *pb.CompositeDoc
	docUrl       *url.URL
	imgUrls      map[string]int
	metaImages   map[string]*pb.ImageElement
	noscriptNode *html.Node
	pictureNode  *html.Node
}

func NewImageExtractor() *ImageExtractor {
	return &ImageExtractor{}
}

func (ie *ImageExtractor) Reset() {
	ie.metaImages = make(map[string]*pb.ImageElement)
	ie.cdoc = &pb.CompositeDoc{
		Images: make([]*pb.ImageElement, 0),
	}
	ie.noscriptNode = nil
	ie.pictureNode = nil
}

func (ie *ImageExtractor) Finalize() {
	glog.V(0).Infof("File has %d images found", len(ie.cdoc.GetImages()))
	glog.V(0).Infof("Composite doc:\n%s", proto.MarshalTextString(ie.cdoc))
	filteredImgEles := make([]*pb.ImageElement, 0)
	ogImgEle := ie.metaImages["og:"]
	if ogImgEle != nil {
	}
	for _, imgEle := range ie.cdoc.Images {
		isOGImg := HasImageSourceGType(imgEle, pb.ImageElement_META_OG)
		if isOGImg {
			utils.IncrementCounterNS("doc", "img_og")
		}
		var nominalUrl string
		if len(imgEle.GetUrl()) > 0 {
			nominalUrl = imgEle.GetUrl()
			glog.V(0).Infof("Nominal image url from src %s", imgEle.GetUrl())
		} else if len(imgEle.GetImageGroups()) == 0 {
			// No srcset. So we pretty much know this is useless
			utils.IncrementCounterNS("doc", "img_meaningless")
			if isOGImg {
				utils.IncrementCounterNS("doc", "img_meaningless_og")
			}
			glog.V(0).Infof("Nominal image url NONE:\n%s",
				proto.MarshalTextString(imgEle))
			continue
		} else {
			// We have srcset either from img tag or from picture tag. But do we have any
			// image url?
			if len(imgEle.GetImageGroups()[0].GetImageSources()) > 0 &&
				len(imgEle.GetImageGroups()[0].GetImageSources()[0].GetUrl()) > 0 {
				nominalUrl = imgEle.GetImageGroups()[0].GetImageSources()[0].GetUrl()
				glog.V(0).Infof("Nominal image url from srcset %s", nominalUrl)
			} else {
				glog.V(0).Infof("Nominal image url NONE from srcset:\n%s",
					proto.MarshalTextString(imgEle))
			}
		}
		if len(nominalUrl) > 0 {
			if ogImgEle != nil && ogImgEle != imgEle && ogImgEle.GetUrl() == nominalUrl {
				glog.V(0).Infof("Merging og ele:%s with img ele:%s",
					ogImgEle.String(), imgEle.String())
				MergeImgEles(ogImgEle, imgEle)
				glog.V(0).Infof("Merged og ele: %s", ogImgEle.String())
			} else {
				filteredImgEles = append(filteredImgEles, imgEle)
			}
		}
	}
	ie.cdoc.Images = filteredImgEles
	utils.IncrementCounterNSBy("doc", "img_extracted", uint32(len(ie.cdoc.GetImages())))
	bucket := len(ie.cdoc.GetImages())
	if bucket >= 10 {
		bucket = bucket - bucket%10
	}
	if bucket >= 100 {
		bucket = bucket - bucket%100
	}
	if bucket >= 10000 {
		bucket = 9999
	}
	utils.IncrementCounterNS("doc", fmt.Sprintf("img_%04d_extracted", uint32(bucket)))
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

	switch n.DataAtom.String() {
	case "img":
		glog.V(1).Infof("Found img tag")
		interested = true
	case "image":
		glog.V(1).Infof("Found image tag")
		utils.IncrementCounterNS("image", "all")
	case "picture":
		if ie.pictureNode != nil {
			glog.Fatal("Found noscript node inside noscript node")
		}
		ie.pictureNode = n
		glog.V(1).Infof("Found picture tag")
		utils.IncrementCounterNS("picture", "all")
		interested = true
	case "noscript":
		embeddedNoscript := false
		if ie.noscriptNode != nil {
			embeddedNoscript = true
		}
		replacedNode, err := ie.ProcessNoscriptNode(n)
		if err != nil {
			glog.Fatal(err)
		}
		if replacedNode != nil {
			ie.ProcessNode(replacedNode)
		}
		if !embeddedNoscript {
			ie.noscriptNode = nil
		}
	case "script":
		typeStr, err := pu.GetAttributeValue(n, "type")
		if err == nil && len(typeStr) > 0 {
			utils.IncrementCounterNS("script", typeStr)
			if typeStr == "application/ld+json" {
				ie.ProcessScriptNode(n)
			}
		}
	case "meta":
		if ie.ProcessMetaNode(n) {
			interested = true
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
		ie.ProcessImgNode(n)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ie.ProcessNode(c)
	}
	if ie.pictureNode == n {
		ie.pictureNode = nil
	}
}

func (ie *ImageExtractor) ProcessImgNode(n *html.Node) {
	utils.IncrementCounterNS("img", "all")
	imgEle := &pb.ImageElement{}
	imgEle.Sources = append(imgEle.Sources, pb.ImageElement_IMG_TAG)
	if ie.noscriptNode != nil {
		imgEle.Sources = append(imgEle.Sources, pb.ImageElement_NOSCRIPT_IMG_TAG)
	}
	if ie.pictureNode != nil {
		imgEle.Sources = append(imgEle.Sources, pb.ImageElement_PICTURE_TAG)
	}
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
	ie.FillImageUrl(n, imgEle)

	glog.V(1).Infof("Getting image element:\n%s",
		proto.MarshalTextString(imgEle))
}

func (ie *ImageExtractor) FillImageUrl(n *html.Node, imgEle *pb.ImageElement) {
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
				glog.Infof("Error in %s", ie.docUrl, urlErr)
				utils.IncrementCounterNS("img", "src-bad")
			} else {
				utils.IncrementCounterNS("img", "src-good")
			}
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
				glog.Info(urlErr)
				utils.IncrementCounterNS("img", "data-src-bad")
			} else {
				utils.IncrementCounterNS("img", "data-src-good")
			}
		} else {
			glog.V(1).Infof("Image Element has empty data-src")
		}
	case pu.ErrAttrNotFound:
	default:
		glog.Fatal(err)
	}

	if srcUrl != nil && dataSrcUrl != nil && srcUrl.String() != dataSrcUrl.String() {
		url = ResolveSrcAndDataSrc(srcUrl, dataSrcUrl)
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
	/*
		var v interface{}
		json.Unmarshal([]byte(n.FirstChild.Data), &v)
		glog.V(1).Infof("Process script data %s", n.FirstChild.Data)
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
	*/
	return nil
}

func (ie *ImageExtractor) ProcessMetaNode(n *html.Node) bool {
	property, err := pu.GetAttributeValue(n, "property")
	if err != nil {
		return false
	}
	if strings.Index(property, "og:") != 0 {
		// Right now we only look at og images
		return false
	}
	metaSrc := "og:"
	imgEle, exist := ie.metaImages[metaSrc]
	if !exist {
		imgEle = &pb.ImageElement{}
		imgEle.Sources = append(imgEle.Sources, pb.ImageElement_META_OG)
		ie.metaImages[metaSrc] = imgEle
		ie.cdoc.Images = append(ie.cdoc.Images, imgEle)
	}
	switch property {
	case "og:image":
		content, err := pu.GetAttributeValue(n, "content")
		if err != nil {
			return false
		}
		if len(content) == 0 {
			glog.V(0).Infof("Nil og content:\n%s", pu.GetLongDisplayNode(n))
		} else {
			imgEle.Url = content
		}
		utils.IncrementCounterNS("meta:og", "imageurl")
	case "og:image:secure_url":
		content, err := pu.GetAttributeValue(n, "content")
		if err != nil {
			return false
		}
		if len(content) == 0 {
			glog.V(0).Infof("Nil og secure url content:\n%s", pu.GetLongDisplayNode(n))
		} else {
			if len(imgEle.Url) == 0 {
				imgEle.Url = content
			}
		}
		utils.IncrementCounterNS("meta:og", "imageurl")
	case "og:image:alt":
		alt, err := pu.GetAttributeValue(n, "content")
		if err != nil {
			return false
		}
		if len(alt) == 0 {
			glog.V(0).Infof("Nil og alt:\n%s", pu.GetLongDisplayNode(n))
		} else {
			imgEle.Alt = alt
		}
	case "og:image:width":
		value, err := pu.GetAttributeIntValue(n, "content")
		if err != nil {
			return false
		}
		imgEle.Width = int32(value)
	case "og:image:height":
		value, err := pu.GetAttributeIntValue(n, "content")
		if err != nil {
			return false
		}
		imgEle.Height = int32(value)
	default:
		utils.IncrementCounterNS("meta:og", property)
		return false
	}
	return true
}

func (ie *ImageExtractor) ProcessNoscriptNode(n *html.Node) (*html.Node, error) {
	if n.DataAtom.String() != "noscript" {
		glog.Fatal("Shouldn't be here, this is just for noscript")
	}
	if n.FirstChild == nil {
		utils.IncrementCounterNS("noscript", "no-child")
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
	imgSrcEles, err := pu.ParseSrcSet(srcset, ie.docUrl)
	if err != nil {
		glog.Info(err)
		utils.IncrementCounterNS("srcset", "abandoned")
		return nil
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

func AddImageSourceType(imgEle *pb.ImageElement, srcType pb.ImageElement_ImageSource) bool {
	if HasImageSourceGType(imgEle, srcType) {
		return false
	}
	imgEle.Sources = append(imgEle.Sources, srcType)
	return true
}

func HasImageSourceGType(imgEle *pb.ImageElement, srcType pb.ImageElement_ImageSource) bool {
	for _, src := range imgEle.Sources {
		if src == srcType {
			return true
		}
	}
	return false
}

func MergeImgEles(dest *pb.ImageElement, src *pb.ImageElement) {
	for _, img := range src.GetImageGroups() {
		dest.ImageGroups = append(dest.ImageGroups, img)
	}
	for _, source := range src.GetSources() {
		AddImageSourceType(dest, source)
	}
}

func ResolveSrcAndDataSrc(srcUrl *url.URL, dataSrcUrl *url.URL) *url.URL {
	glog.V(1).Infof("Trying to pick from src %s and data-src %s",
		srcUrl.String(), dataSrcUrl.String())
	utils.IncrementCounterNS("img", "src-ambiguous")
	glog.V(1).Infof("Scheme %s vs. %s", srcUrl.Scheme, dataSrcUrl.Scheme)
	if srcUrl.Scheme == "data" && dataSrcUrl.Scheme != "data" {
		utils.IncrementCounterNS("img", "src-resolved-by-scheme")
		return dataSrcUrl
	} else if GetImageFileType(dataSrcUrl) == ImageFileType_JPG &&
		GetImageFileType(srcUrl) != ImageFileType_JPG {
		utils.IncrementCounterNS("img", "src-resolved-by-ext")
		return dataSrcUrl
	}
	glog.V(1).Infof("Unresolved src %s and data-src %s, with path %s vs. %s",
		srcUrl.String(), dataSrcUrl.String(),
		srcUrl.Path, dataSrcUrl.Path)
	utils.IncrementCounterNS("img", "src-unresolved")
	return nil
}

func GetImageFileType(srcUrl *url.URL) ImageFileType {
	ext := GetImageExt(srcUrl)
	fileType := GetImageFileTypeByExt(ext)
	return fileType
}

func GetImageFileTypeByExt(ext string) ImageFileType {
	switch ext {
	case ".jpeg":
		return ImageFileType_JPG
	case ".jpg":
		return ImageFileType_JPG
	case ".png":
		return ImageFileType_PNG
	case ".gif":
		return ImageFileType_GIF
	default:
		return ImageFileType_UNKNOWN
	}
}

func GetImageExt(srcUrl *url.URL) string {
	ext := strings.ToLower(filepath.Ext(srcUrl.Path))
	return ext
}
