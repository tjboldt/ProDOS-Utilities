package prodos

import (
	"os"
)

func ReadBlock(file *os.File, block int) []byte {
	buffer := make([]byte, 512)

	file.ReadAt(buffer, int64(block)*512)

	return buffer
}
