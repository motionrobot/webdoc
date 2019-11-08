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
)

var (
	serpScrapeInfoFilePtr = flag.String(
		"serp_scrape_info_file",
		"",
		"The input file")
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
			var matchedImgEle *pb.ImageElement
			for _, imgEle := range cdoc.GetImages() {
				if imgEle.GetUrl() == result.GetImageUrl() {
					matchedImgEle = imgEle
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
