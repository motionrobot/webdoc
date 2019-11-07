package main

import (
	"bufio"
	"flag"
	"github.com/golang/glog"
	"github.com/motionrobot/webdoc/extractor"
	pu "github.com/motionrobot/webdoc/parserutils"
	"github.com/motionrobot/webdoc/serp"
	"os"
	"path/filepath"
	"strings"
)

var (
	inputFilesPtr = flag.String(
		"input_files",
		"",
		"The patterns of the local files")
	dataDirPtr = flag.String(
		"data_dir",
		"",
		"The directory of the local files")
	parsersPtr = flag.String(
		"parsers",
		"",
		"comma separated list of parser names")
)

func main() {
	flag.Parse()

	if len(*parsersPtr) == 0 {
		glog.Fatal("No parsers specified")
	}
	parserNames := strings.Split(*parsersPtr, ",")
	glog.Infof("Processing with parsers %v", parserNames)
	parsers := make([]pu.Parser, len(parserNames))
	for i, parserName := range parserNames {
		var a pu.Parser
		switch parserName {
		case "ImageExtractor":
			a = extractor.NewImageExtractor()
		case "SERPParser":
			a = serp.NewSERPParser()
		default:
			glog.Fatal("Unknown parser name", parserName)
		}
		glog.V(0).Infof("Getting hander %T from name %s", a, parserName)
		parsers[i] = a
	}

	var files []string
	var err error
	if len(*inputFilesPtr) > 0 {
		files, err = filepath.Glob(*inputFilesPtr)
		if err != nil {
			glog.Fatal(err)
		}
	}
	glog.Infof("Parsing %d files", len(files))
	for _, fn := range files {
		glog.V(1).Infof("%s", fn)
	}
	for _, fn := range files {
		glog.V(0).Infof("Processing file %s", fn)
		f, err := os.Open(fn)
		if err != nil {
			glog.Fatal(err)
		}
		reader := bufio.NewReader(f)
		for _, p := range parsers {
			p.Reset()
			if err := p.Parse(reader, nil); err != nil {
				glog.Fatal(err)
			}
			p.Finalize()
		}
	}
}
