package main

import (
	"encoding/base64"
	"flag"
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/motionrobot/utils"
	"github.com/motionrobot/webdoc/extractor"
	pu "github.com/motionrobot/webdoc/parserutils"
	pb "github.com/motionrobot/webdoc/proto"
	"net/url"
)

var (
	serpScrapeInfoFilePtr = flag.String(
		"serp_scrape_info_file",
		"",
		"The input file")
	debugDocUrlPtr = flag.String(
		"debug_doc_url",
		"",
		"The url to be debugged")
)

func main() {
	flag.Parse()

	if len(*serpScrapeInfoFilePtr) == 0 {
		glog.Fatal("No serp scraper file")
	}
	scrapeInfos := make([]*pb.SERPScrapeInfo, 0)
	slr := utils.NewSimpleLineReaderWithCallback(
		*serpScrapeInfoFilePtr,
		func(line string) bool {
			base64_decoded, err := base64.StdEncoding.DecodeString(line)
			if err != nil {
				glog.Fatal(err)
			}
			result := &pb.SERPScrapeInfo{}
			err = proto.Unmarshal(base64_decoded, result)
			if err != nil {
				glog.Fatal(err)
			}
			scrapeInfos = append(scrapeInfos, result)
			return true
		})
	slr.ProcessLines()

	ie := extractor.NewImageExtractor()
	cdocs := make(map[string]*pb.CompositeDoc)
	for _, scrapeInfo := range scrapeInfos {
		for _, result := range scrapeInfo.GetResultPage().GetResults() {
			if len(*debugDocUrlPtr) > 0 && *debugDocUrlPtr != result.GetUrl() {
				continue
			}
			fn, exist := scrapeInfo.GetCachedFiles()[result.GetPos()]
			if !exist {
				utils.IncrementCounterNS("doc", "MissingFile")
				glog.Info("Cached file not found for %d", result.GetPos())
				continue
			}
			glog.V(0).Infof("Processing cached file %s log result:\n%s",
				fn, proto.MarshalTextString(result))
			ie.Reset()
			cdoc := &pb.CompositeDoc{}
			cdoc.Url = result.GetUrl()
			cdocs[cdoc.GetUrl()] = cdoc
			if err := pu.ParseFile(fn, ie, cdoc); err != nil {
				glog.Fatal(err)
			}
			ie.Finalize()
		}
	}
	for _, scrapeInfo := range scrapeInfos {
		for _, result := range scrapeInfo.GetResultPage().GetResults() {
			if len(result.GetImageUrl()) == 0 {
				utils.IncrementCounterNS("result", "no-image-url")
				continue
			}
			cdoc, exist := cdocs[result.GetUrl()]
			if !exist {
				continue
			}
			glog.V(0).Infof("Composite doc from cached file %s:\n%s",
				scrapeInfo.GetCachedFiles()[result.GetPos()],
				proto.MarshalTextString(cdoc))
			imgUrl, err := url.Parse(result.GetImageUrl())
			if err != nil {
				glog.Fatal(err)
			}
			var matchedImgEle *pb.ImageElement
			for _, imgEle := range cdoc.GetImages() {
				if UrlMatches(imgUrl, imgEle.GetUrl()) {
					matchedImgEle = imgEle
					break
				}
				for _, group := range imgEle.GetImageGroups() {
					for _, srcEle := range group.GetImageSources() {
						if UrlMatches(imgUrl, srcEle.GetUrl()) {
							matchedImgEle = imgEle
							utils.IncrementCounterNS("result", "image-url-srcset-matched")
							break
						}
					}
					if matchedImgEle != nil {
						break
					}
				}
				if matchedImgEle != nil {
					break
				}
			}
			if matchedImgEle != nil {
				glog.V(0).Infof("Image Found:%s",
					proto.MarshalTextString(matchedImgEle))
				utils.IncrementCounterNS("result", "image-url-found")
			} else {
				utils.IncrementCounterNS("result", "image-url-missing")
				glog.V(0).Infof("Image Missing:%s", result.GetImageUrl())
			}
		}
	}
	utils.PrintCounters()
}

func UrlMatches(imgUrl *url.URL, urlStr string) bool {
	if imgUrl.String() == urlStr {
		return true
	}
	newUrl, err := url.Parse(urlStr)
	if err != nil {
		glog.Fatal(err)
	}
	if newUrl.Host == imgUrl.Host && newUrl.Path == imgUrl.Path {
		glog.V(0).Infof("Matching url after dropping query seg: %s vs %s (%s %s) vs (%s %s)",
			imgUrl.String(), urlStr, imgUrl.Host, imgUrl.Path, newUrl.Host, newUrl.Path)
		return true
	}
	return false
}
