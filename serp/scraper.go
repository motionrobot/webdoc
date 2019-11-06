package serp

import (
	"compress/gzip"
	"flag"
	"fmt"
	"github.com/golang/glog"
	pb "github.com/motionrobot/webdoc/proto"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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

var (
	scrapeOutputFilePtr = flag.String(
		"scrape_output_file",
		"",
		"The output file")
)

type SERPScrapeResult struct {
	ResultPage  *pb.GenericSearchResponse
	CachedFiles map[uint32]string
}

type SERPScraper struct {
	serpParser *SERPParser
	client     *http.Client
	result     *SERPScrapeResult
}

func NewSERPScraper() *SERPScraper {
	return &SERPScraper{
		serpParser: NewSERPParser(),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		result: &SERPScrapeResult{},
	}
}

func (s *SERPScraper) ScrapeSERPFile(fn string) *SERPScrapeResult {
	s.result = &SERPScrapeResult{CachedFiles: make(map[uint32]string)}
	s.serpParser.Reset()
	defer s.serpParser.Finalize()
	if err := s.serpParser.ParseFile(fn); err != nil {
		glog.Fatal(err)
	}
	s.result.ResultPage = s.serpParser.GetResultPage()
	baseName, err := filepath.Abs(fn)
	if err != nil {
		glog.Fatal(err)
	}
	ext := filepath.Ext(baseName)
	if len(ext) > 0 {
		baseName = baseName[:strings.LastIndex(baseName, ext)]
	}

	for _, result := range s.result.ResultPage.GetResults() {
		data, err := s.ScrapeResult(result)
		if err != nil || data == nil {
			continue
		}
		tmpFileName := fmt.Sprintf("%s_result_%d%s", baseName, result.Pos, ext)
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
		s.result.CachedFiles[result.GetPos()] = tmpFileName
	}
	return s.result
}

func (s *SERPScraper) ScrapeResult(result *pb.Result) ([]byte, error) {
	glog.V(0).Infof("Scraping %s", result.Url)
	request, err := http.NewRequest("GET", result.Url, nil)
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36")
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
