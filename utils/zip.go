package utils

import (
	"bytes"
	"compress/gzip"
	"io"
)

// Unzip unzips the input byte slice.
func Unzip(in []byte) ([]byte, error) {
	inReader := bytes.NewReader(in)
	gzipReader, err := gzip.NewReader(inReader)
	if err != nil {
		return nil, err
	}
	defer gzipReader.Close()
	var result bytes.Buffer
	_, err = result.ReadFrom(gzipReader)
	if err != nil {
		return nil, err
	}
	err = gzipReader.Close()
	if err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}

// Zip zips the input byte slice with the given gzip compression level.
func Zip(in []byte, level int) ([]byte, error) {
	var result bytes.Buffer
	inReader := bytes.NewReader(in)
	gzipWriter, err := gzip.NewWriterLevel(&result, level)
	if err != nil {
		return nil, err
	}
	defer gzipWriter.Close()
	_, err = io.Copy(gzipWriter, inReader)
	if err != nil {
		return nil, err
	}
	err = gzipWriter.Close()
	if err != nil {
		return nil, err
	}
	return result.Bytes(), nil
}
