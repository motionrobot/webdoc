package serp

import (
	"github.com/golang/glog"
	"testing"
)

func TestImageSearchScraper(t *testing.T) {
	files := []string{
		"/Users/zheng/work/data/scraper/srp1.html",
		"/Users/zheng/work/data/scraper/srp2.html",
	}
	p := NewSERPScraper()
	defer p.Close()
	for _, fn := range files {
		glog.Infof("Testing file %s", fn)
		p.ScrapeSERPFile(fn)
	}
}
