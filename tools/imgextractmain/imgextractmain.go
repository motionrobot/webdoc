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
	cdocs := make([]*pb.CompositeDoc, 0)
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
			cdocs = append(cdocs, cdoc)
			if err := pu.ParseFile(fn, ie, cdoc); err != nil {
				glog.Fatal(err)
			}
			ie.Finalize()
		}
	}
	for _, cdoc := range cdocs {
		glog.V(0).Infof("All composite docs:\n%s", proto.MarshalTextString(cdoc))
	}
	utils.PrintCounters()
}
