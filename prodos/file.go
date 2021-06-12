package prodos

import (
	"os"
	"strings"
)

func LoadFile(file *os.File, path string) []byte {
	fileEntry := GetFileEntry(file, path)

	blockList := GetBlocklist(file, fileEntry)

	buffer := make([]byte, fileEntry.EndOfFile)

	for i := 0; i < len(blockList); i++ {
		block := ReadBlock(file, blockList[i])
		for j := 0; j < 512 && i*512+j < fileEntry.EndOfFile; j++ {
			buffer[i*512+j] = block[j]
		}
	}

	return buffer
}

func GetBlocklist(file *os.File, fileEntry FileEntry) []int {
	blocks := make([]int, fileEntry.BlocksUsed)

	switch fileEntry.StorageType {
	case StorageSeedling:
		blocks[0] = fileEntry.KeyPointer
	case StorageSapling:
		index := ReadBlock(file, fileEntry.KeyPointer)
		for i := 0; i < fileEntry.BlocksUsed-1; i++ {
			blocks[i] = int(index[i]) + int(index[i+256])*256
		}
	case StorageTree:
		masterIndex := ReadBlock(file, fileEntry.KeyPointer)
		for i := 0; i < 128; i++ {
			index := ReadBlock(file, int(masterIndex[i])+int(masterIndex[i+256])*256)
			for j := 0; j < 256 && i*256+j < fileEntry.BlocksUsed; j++ {
				blocks[i*256+j] = int(index[j]) + int(index[j+256])*256
			}
		}
	}

	return blocks
}

func GetFileEntry(file *os.File, path string) FileEntry {
	path = strings.ToUpper(path)
	paths := strings.Split(path, "/")

	var directoryBuilder strings.Builder

	for i := 1; i < len(paths)-1; i++ {
		directoryBuilder.WriteString("/")
		directoryBuilder.WriteString(paths[i])
	}

	directory := directoryBuilder.String()
	fileName := paths[len(paths)-1]

	_, fileEntries := ReadDirectory(file, directory)

	if fileEntries == nil {
		return FileEntry{}
	}

	var fileEntry FileEntry

	for i := 0; i < len(fileEntries); i++ {
		if fileEntries[i].FileName == fileName {
			fileEntry = fileEntries[i]
		}
	}

	return fileEntry
}

func DeleteFile(file *os.File, path string) {
	fileEntry := GetFileEntry(file, path)

	// free the blocks
	blocks := GetBlocklist(file, fileEntry)
	volumeBitmap := ReadVolumeBitmap(file)
	for i := 0; i < len(blocks); i++ {
		FreeBlockInVolumeBitmap(volumeBitmap, blocks[i])
	}
	WriteVolumeBitmap(file, volumeBitmap)

	// zero out directory entry
	fileEntry.StorageType = 0
	fileEntry.FileName = ""
	writeFileEntry(file, fileEntry)

	// decrement the directory entry count
	directoryBlock := ReadBlock(file, fileEntry.HeaderPointer)
	directoryHeader := parseDirectoryHeader(directoryBlock)

	directoryHeader.ActiveFileCount--
	writeDirectoryHeader(file, directoryHeader, fileEntry.HeaderPointer)
}
