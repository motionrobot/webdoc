package serp

import (
	"github.com/golang/glog"
	"testing"
)

func TestImageSearchParser(t *testing.T) {
	files := []string{"/home/zheng/work/data/srp2.html"}
	for _, fn := range files {
		glog.Infof("Testing file %s", fn)
		p := NewSERPParser()
		p.Reset()
		p.ParseFile(fn)
		p.Finalize()
	}
}
