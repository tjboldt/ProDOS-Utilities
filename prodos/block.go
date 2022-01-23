// Copyright Terence J. Boldt (c)2021-2022
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides access to read and write
// blocks on a ProDOS drive image

package prodos

import (
	"io"
)

func ReadBlock(reader io.ReaderAt, block int) []byte {
	buffer := make([]byte, 512)

	reader.ReadAt(buffer, int64(block)*512)

	return buffer
}

func WriteBlock(writer io.WriterAt, block int, buffer []byte) {
	writer.WriteAt(buffer, int64(block)*512)
}
