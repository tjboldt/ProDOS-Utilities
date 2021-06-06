package prodos

import (
	"fmt"
	"os"
)

func WriteBlock(file *os.File, block int, buffer []byte) {

	fmt.Printf("Write block %d\n", block)

	file.WriteAt(buffer, int64(block)*512)
	file.Sync()
	fmt.Printf("Write block completed\n")
}
