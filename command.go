package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	// default size of a splitted part in bytes
	defaultSize = 20 * 1024 * 1024
)

// Split splits a big file into many parts.
// The maximal size of each part is size bytes.
func Split(fileName string, size int, outputFileName string) (parts []string, err error) {
	if fileName == "" {
		return nil, errors.New("file name is empty")
	}

	if size == 0 {
		size = defaultSize
	}

	if outputFileName == "" {
		return nil, errors.New("empty output file name")
	}

	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("%s does not exist", fileName)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("%s is a directory, not a file", fileName)
	}

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

// Merge merges files into a big file by simply concatenating them one by one
func Merge(fileNames []string, outputFileName string) (err error) {
	if len(fileNames) == 0 {
		return errors.New("empty file list")
	}

	if outputFileName == "" {
		return errors.New("empty output file name")
	}

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

