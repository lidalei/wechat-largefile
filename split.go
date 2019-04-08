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
func Split(fileName string, size int, outputFileName string) (err error) {
	if fileName == "" {
		return errors.New("file name is empty")
	}

	if size == 0 {
		size = defaultSize
	}

	if outputFileName == "" {
		outputFileName = fileName
	}

	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist", fileName)
	}

	if info.IsDir() {
		return fmt.Errorf("%s is a directory, not a file", fileName)
	}

	f, err := os.Open(fileName)
	if err != nil {
		return err
	}

	// check err returned from closing a file.
	defer func() {
		err = f.Close()
	}()

	buf := make([]byte, size)
	for i := 1; ; i++ {
		l, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// write to a file
		err = write(fmt.Sprintf("%s.part%d", outputFileName, i), buf[:l])
		if err != nil {
			return err
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// Merge merges files into a big file by simply concatenating them one by one
func Merge(fileNames []string, outputFileName string) (err error) {
	if len(fileNames) == 0 {
		return errors.New("empty file list")
	}

	if outputFileName == "" {
		return errors.New("empty output file name")
	}

	// read file one and write one

	return nil
}

func write(fileName string, data []byte) (err error) {
	fmt.Println("shit")
	if fileName == "" {
		return errors.New("file name is empty")
	}

	_, err = os.Stat(fileName)
	if os.IsExist(err) {
		return fmt.Errorf("%s exists", fileName)
	}

	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0755)
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

