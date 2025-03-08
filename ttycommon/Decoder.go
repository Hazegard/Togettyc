package ttycommon

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/klauspost/compress/zstd"
	"io"
	"os"
	"togettyc/ttyrec"
)

var zstdMagic = []byte{0x28, 0xB5, 0x2F, 0xFD}

func NewDecoder(recordFile string) (*ttyrec.Decoder, []io.ReadCloser, error) {
	openedFiles := make([]io.ReadCloser, 0)
	var inputFile io.ReadCloser
	if recordFile == "" {
		inputFile = os.Stdin
	} else {
		var err error
		if inputFile, err = os.Open(recordFile); err != nil {
			return nil, nil, fmt.Errorf("error opening %s: %v\n", recordFile, err)
		}
		openedFiles = append(openedFiles, inputFile)
	}

	// Wrap the original reader in a bufio.Reader to peek at magic bytes.
	br := bufio.NewReader(inputFile)
	peek, err := br.Peek(4)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading magic bytes: %v\n", err)
	}

	var decoder *ttyrec.Decoder
	if bytes.Equal(peek, zstdMagic) {
		// Create a zstd decoder using the buffered reader.
		zstdDecoder, err := zstd.NewReader(br)
		if err != nil {
			return nil, nil, fmt.Errorf("error creating zstd reader: %v\n", err)
		}
		zstdReader := zstdDecoder.IOReadCloser()
		openedFiles = append(openedFiles, zstdReader)
		decoder = ttyrec.NewDecoder(zstdReader)
	} else {
		decoder = ttyrec.NewDecoder(br)
	}
	return decoder, openedFiles, nil
}
