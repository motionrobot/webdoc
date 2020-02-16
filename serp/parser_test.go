package serp

import (
	"github.com/golang/glog"
	pu "github.com/motionrobot/webdoc/parserutils"
	"testing"
)

func TestImageSearchParser(t *testing.T) {
	files := []string{
		"/Users/zhengsun/work/data/image/explicit_image_safari/serp/scraped_imageSearch_desktop_en-US_0000000.html",
		"/Users/zhengsun/work/data/image/explicit_image_safari/serp/scraped_imageSearch_desktop_en-US_0000001.html",
	}
	p := NewSERPParser()
	for _, fn := range files {
		glog.Infof("Testing file %s", fn)
		p.Reset()
		pu.ParseFile(fn, p, nil)
		p.Finalize()
	}
}
