package serp

import (
	"github.com/golang/glog"
	"testing"
)

func TestImageSearchScraper(t *testing.T) {
	files := []string{
		"/Users/zhengsun/work/data/image/explicit_image_safari/serp/scraped_imageSearch_desktop_en-US_0000000.html",
		"/Users/zhengsun/work/data/image/explicit_image_safari/serp/scraped_imageSearch_desktop_en-US_0000001.html",
	}
	p := NewSERPScraper()
	defer p.Close()
	for _, fn := range files {
		glog.Infof("Testing file %s", fn)
		p.ScrapeSERPFile(fn)
	}
}
