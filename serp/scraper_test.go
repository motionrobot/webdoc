package serp

import (
	"github.com/golang/glog"
	"testing"
)

func TestImageSearchScraper(t *testing.T) {
	files := []string{
		"/Users/zhengsun/work/data/imagescraper/srp1.html",
		"/Users/zhengsun/work/data/imagescraper/srp2.html",
	}
	p := NewSERPScraper()
	defer p.Close()
	for _, fn := range files {
		glog.Infof("Testing file %s", fn)
		p.ScrapeSERPFile(fn)
	}
}
