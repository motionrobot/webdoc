package serp

import (
	"compress/gzip"
	"fmt"
	"github.com/golang/glog"
	pb "github.com/motionrobot/webdoc/proto"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
	/*
		"bufio"
		"encoding/json"
		"github.com/golang/protobuf/proto"
		pu "github.com/motionrobot/webdoc/parserutils"
		"golang.org/x/net/html"
		"strconv"
	*/)

type SERPScraper struct {
	serpParser *SERPParser
	client     *http.Client
	//resultParser *ResultParser
}

func NewSERPScraper() *SERPScraper {
	return &SERPScraper{
		serpParser: NewSERPParser(),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *SERPScraper) ScrapeSERPFile(fn string) {
	resultPage := s.serpParser.ParseFile(fn)
	for _, result := range resultPage.GetResults() {
		data, err := s.ScrapeResult(result)
		if err != nil || data == nil {
			continue
		}
		tmpFileName := fmt.Sprintf("%s_%d", fn, result.Pos)
		glog.V(0).Infof("Writing to file %s", tmpFileName)
		tmpFile, err := os.Create(tmpFileName)
		if err != nil {
			glog.Fatal(err)
		}
		n, err := tmpFile.Write(data)
		if err != nil {
			glog.Fatal(err)
		}
		if n != len(data) {
			glog.Fatal("Data written wrong size")
		}
		tmpFile.Close()
	}
}

func (s *SERPScraper) ScrapeResult(result *pb.Result) ([]byte, error) {
	glog.V(0).Infof("Scraping %s", result.Url)
	request, err := http.NewRequest("GET", result.Url, nil)
	if err != nil {
		glog.Infof("Failed to build request %s: %v", result.Url, err)
	}
	resp, err := s.client.Do(request)
	if err != nil {
		glog.Infof("Failed to load %s: %v", result.Url, err)
		return nil, err
	}
	defer resp.Body.Close()
	glog.V(0).Infof("Response %d", resp.StatusCode)
	if resp.StatusCode >= 400 {
		return nil, nil
	}
	var bodyReader io.Reader = resp.Body
	contentEncoding := strings.ToLower(resp.Header.Get("Content-Encoding"))
	if !resp.Uncompressed && (strings.Contains(contentEncoding, "gzip") || (contentEncoding == "" && strings.Contains(strings.ToLower((resp.Header.Get("Content-Type"))), "gzip"))) {
		bodyReader, err = gzip.NewReader(bodyReader)
		if err != nil {
			return nil, err
		}
		defer bodyReader.(*gzip.Reader).Close()
	}
	body, err := ioutil.ReadAll(bodyReader)
	return body, nil
}
