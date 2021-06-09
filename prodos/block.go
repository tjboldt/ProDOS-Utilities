package prodos

import (
	"os"
)

func ReadBlock(file *os.File, block int) []byte {
	buffer := make([]byte, 512)

	file.ReadAt(buffer, int64(block)*512)

	return buffer
}

func WriteBlock(file *os.File, block int, buffer []byte) {
	WriteBlockNoSync(file, block, buffer)
	file.Sync()
}

func WriteBlockNoSync(file *os.File, block int, buffer []byte) {
	file.WriteAt(buffer, int64(block)*512)
}
