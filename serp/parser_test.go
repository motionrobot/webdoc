package serp

import (
	"github.com/golang/glog"
	"testing"
)

func TestImageSearchResult(t *testing.T) {
	files := []string{"/home/zheng/work/data/srp2.html"}
	for _, fn := range files {
		glog.Infof("Testing file %s", fn)
		p := NewSRPParser()
		p.ParseFile(fn)
	}
}
