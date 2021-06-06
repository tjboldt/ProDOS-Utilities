package prodos

import (
	"fmt"
	"os"
	"strings"
)

func LoadFile(file *os.File, path string) []byte {
	path = strings.ToUpper(path)
	paths := strings.Split(path, "/")

	var directoryBuilder strings.Builder

	for i := 1; i < len(paths)-1; i++ {
		directoryBuilder.WriteString("/")
		directoryBuilder.WriteString(paths[i])
	}

	directory := directoryBuilder.String()
	fileName := paths[len(paths)-1]

	volumeHeader, fileEntries := ReadDirectory(file, directory)

	if fileEntries == nil {
		return nil
	}

	var fileEntry FileEntry

	for i := 0; i < len(fileEntries); i++ {
		if fileEntries[i].FileName == fileName {
			fileEntry = fileEntries[i]
		}
	}

	DumpVolumeHeader(volumeHeader)

	fmt.Println()

	DumpFileEntry(fileEntry)

	switch fileEntry.StorageType {
	case StorageSeedling:
		return ReadBlock(file, fileEntry.StartingBlock)[0:fileEntry.EndOfFile]
	case StorageSapling:
		index := ReadBlock(file, fileEntry.StartingBlock)
		buffer := make([]byte, fileEntry.EndOfFile)
		for i := 0; i < 512 && index[i] > 0; i++ {
			chunk := ReadBlock(file, int(index[i])+int(index[i+256])*256)
			for j := i * 512; j < fileEntry.EndOfFile && j < i*512+512; j++ {
				buffer[j] = chunk[j-i*512]
			}
		}
		return buffer
	case StorageTree:
		// add tree file support later
		return nil
	}
	return nil
}
