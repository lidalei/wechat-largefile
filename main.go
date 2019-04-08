package main

import (
	"flag"

	"github.com/prometheus/common/log"
)

var (
	fileName       = flag.String("file", "", "The big file to split")
	size           = flag.Int("size", 20, "The maximal size per part in megabytes")
	outputFileName = flag.String("output", "", "The prefix of output files, the same to file by default")
)

func main() {
	flag.Parse()

	// split file
	err := Split(*fileName, *size*1024*1024, *outputFileName)
	if err != nil {
		log.Errorf("fail to split file %s, error: %v", *fileName, err)
	}
}

