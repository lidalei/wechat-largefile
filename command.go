package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/subcommands"
	"github.com/sirupsen/logrus"
)

const (
	// default size of a splitted part in MegaBytes
	defaultSize = 20
)

// splitCmd implements subcommands.Command interface
type splitCmd struct {
	fileName       string
	size           int
	outputFileName string
}

// Name returns the name of the command.
func (sc *splitCmd) Name() string {
	return "split"
}

// Synopsis returns a short string (less than one line) describing the command.
func (sc *splitCmd) Synopsis() string {
	return "split splits a big file into smaller parts"
}

// Usage returns a long string explaining the command and giving usage
// information.
func (sc *splitCmd) Usage() string {
	return `split -file=FileName [-size=Size] [-out=FileName].
`
}

// SetFlags adds the flags for this command to the specified set.
func (sc *splitCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&sc.fileName, "file", "", "The big file to split")
	flag.IntVar(&sc.size, "size", defaultSize, "The maximal size per part in megabytes")
	flag.StringVar(&sc.outputFileName, "out", "", "The prefix of output files, the same to file by default")
}

// Execute executes the command and returns an ExitStatus.
func (sc *splitCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) (status subcommands.ExitStatus) {
	var err error
	var parts []string
	defer func() {
		if err != nil {
			logrus.Errorf("failed to split file %s, error: %v", sc.fileName, err)
		} else {
			logrus.Infof("split file %s into parts: %v", sc.fileName, parts)
		}
	}()
	// Validate arguments and set them with default values if necessary
	if sc.fileName == "" {
		err = errors.New("file name is empty")
		status = subcommands.ExitUsageError
		return
	}

	if sc.size <= 0 {
		err = errors.New("size must be positive")
		status = subcommands.ExitUsageError
		return
	}
	// convert MegaBytes to Bytes
	sc.size *= 1024 * 1024

	if sc.outputFileName == "" {
		sc.outputFileName = sc.fileName
	}

	info, err := os.Stat(sc.fileName)
	if os.IsNotExist(err) {
		err = fmt.Errorf("%s does not exist", sc.fileName)
		status = subcommands.ExitUsageError
		return
	}

	if info.IsDir() {
		err = fmt.Errorf("%s is a directory, not a file", sc.fileName)
		status = subcommands.ExitUsageError
		return
	}

	parts, err = split(sc.fileName, sc.size, sc.outputFileName)
	if err != nil {
		status = subcommands.ExitFailure
	} else {
		status = subcommands.ExitSuccess
	}
	return
}

// Split splits a big file into many parts.
// The maximal size of each part is size bytes.
func split(fileName string, size int, outputFileName string) (parts []string, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	// check err returned from closing a file.
	defer func() {
		err = f.Close()
	}()

	buf := make([]byte, size)
	parts = make([]string, 0, 2)
	for i := 1; ; i++ {
		l, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// write to a file
		part := fmt.Sprintf("%s.part%d", outputFileName, i)
		parts = append(parts, part)
		err = write(part, buf[:l], false)
		if err != nil {
			return nil, err
		}

		if err != nil {
			return nil, err
		}
	}

	return
}

// mergeCmd implements subcommands.Command interface
// FIMXE! Support file glob!
type mergeCmd struct {
	fileNames      string
	outputFileName string
}

// Name returns the name of the command.
func (mc *mergeCmd) Name() string {
	return "merge"
}

// Synopsis returns a short string (less than one line) describing the command.
func (mc *mergeCmd) Synopsis() string {
	return "merge smaller files (usually obtained from the counter-command split) into a big file"
}

// Usage returns a long string explaining the command and giving usage
// information.
func (mc *mergeCmd) Usage() string {
	return `merge -files=fileNames -out=fileName.
`
}

// SetFlags adds the flags for this command to the specified set.
func (mc *mergeCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&mc.fileNames, "files", "", "File names separated by ,")
	f.StringVar(&mc.outputFileName, "out", "", "Output file name")
}

// Execute executes the command and returns an ExitStatus.
func (mc *mergeCmd) Execute(ctx context.Context, f *flag.FlagSet, args ...interface{}) (status subcommands.ExitStatus) {
	var err error
	var fileNames []string
	defer func() {
		if err != nil {
			logrus.Errorf("failed to merge files %v, error: %v", fileNames, err)
		} else {
			logrus.Infof("merged parts %v into file %s", fileNames, mc.outputFileName)
		}
	}()

	// Parse and validate arguments
	fileNames = strings.Split(mc.fileNames, ",")
	if len(fileNames) == 0 {
		err = errors.New("empty file list")
		status = subcommands.ExitUsageError
		return
	}

	if mc.outputFileName == "" {
		err = errors.New("empty output file name")
		status = subcommands.ExitUsageError
		return
	}
	err = merge(fileNames, mc.outputFileName)
	if err != nil {
		status = subcommands.ExitFailure
	} else {
		status = subcommands.ExitSuccess
	}
	return
}

// merge merges files into a big file by simply concatenating them one by one
func merge(fileNames []string, outputFileName string) (err error) {
	// create a file to write to
	err = write(outputFileName, []byte{}, false)
	if err != nil {
		return fmt.Errorf("fail to create file %s, error: %v", outputFileName, err)
	}

	// read file one by one and write to one
	for _, fileName := range fileNames {
		data, err := read(fileName)
		if err != nil {
			return err
		}

		err = write(outputFileName, data, true)
		if err != nil {
			return err
		}
	}

	return nil
}

func read(fileName string) (data []byte, err error) {
	if fileName == "" {
		return nil, errors.New("file name is empty")
	}

	_, err = os.Stat(fileName)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("%s does not exist", fileName)
	}

	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	// check err returned from closing a file.
	defer func() {
		err = f.Close()
	}()

	data = make([]byte, 0, defaultSize)

	buf := make([]byte, defaultSize)
	for {
		l, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		data = append(data, buf[:l]...)
	}

	return
}

// write writes data into fileName in an append mode.
// If fileName does not exist, it will be created.
func write(fileName string, data []byte, isAppend bool) (err error) {
	if fileName == "" {
		return errors.New("file name is empty")
	}

	writeFlag := os.O_WRONLY | os.O_CREATE
	if isAppend {
		writeFlag = writeFlag | os.O_APPEND
	}

	f, err := os.OpenFile(fileName, writeFlag, 0644)
	if err != nil {
		return err
	}

	// check err returned from closing a file.
	defer func() {
		err = f.Close()
	}()

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return
}

