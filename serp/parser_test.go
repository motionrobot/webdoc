package serp

import (
	"github.com/golang/glog"
	pu "github.com/motionrobot/webdoc/parserutils"
	"testing"
)

func TestImageSearchParser(t *testing.T) {
	files := []string{
		"/Users/zhengsun/work/data/imagescraper/srp1.html",
		"/Users/zhengsun/work/data/imagescraper/srp2.html",
	}
	p := NewSERPParser()
	for _, fn := range files {
		glog.Infof("Testing file %s", fn)
		p.Reset()
		pu.ParseFile(fn, p, nil)
		p.Finalize()
	}
}
