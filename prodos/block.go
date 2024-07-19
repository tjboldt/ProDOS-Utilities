// Copyright Terence J. Boldt (c)2021-2024
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides access to read and write
// blocks on a ProDOS drive image

package prodos

import (
	"errors"
	"fmt"
	"io"
)

// ReadBlock reads a block from a ProDOS volume into a byte array
func ReadBlock(reader io.ReaderAt, block uint16) ([]byte, error) {
	buffer := make([]byte, 512)

	_, err := reader.ReadAt(buffer, int64(block)*512)
	if err != nil {
		errString := fmt.Sprintf("failed to read block %04X: %s", block, err.Error())
		err = errors.New(errString)
	}

	return buffer, err
}

// WriteBlock writes a block to a ProDOS volume from a byte array
func WriteBlock(writer io.WriterAt, block uint16, buffer []byte) error {
	_, err := writer.WriteAt(buffer, int64(block)*512)
	if err != nil {
		errString := fmt.Sprintf("failed to write block %04X: %s", block, err.Error())
		err = errors.New(errString)
	}

	return err
}
