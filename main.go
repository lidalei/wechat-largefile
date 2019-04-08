package main

import (
	"flag"

	"github.com/prometheus/common/log"
	// "github.com/google/subcommands"
)

var (
	fileName       = flag.String("file", "", "The big file to split")
	size           = flag.Int("size", 20, "The maximal size per part in megabytes")
	outputFileName = flag.String("output", "", "The prefix of output files, the same to file by default")
)

func main() {
	flag.Parse()

	if *outputFileName == "" {
		*outputFileName = *fileName
	}

	// split file
	parts, err := Split(*fileName, *size*1024*1024, *outputFileName)
	if err != nil {
		log.Errorf("fail to split file %s, error: %v", *fileName, err)
	}

	log.Infof("write to parts: %v", parts)

	// merge files
	o := *outputFileName + ".new"
	err = Merge(parts, o)
	if err != nil {
		log.Errorf("fail to merge files %v, error: %v", parts, err)
	}
	log.Infof("merge parts %v into file %s", parts, o)
}

