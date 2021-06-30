package prodos

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

func LoadFile(file *os.File, path string) ([]byte, error) {
	fileEntry, err := GetFileEntry(file, path)
	if err != nil {
		return nil, err
	}

	blockList := getBlocklist(file, fileEntry)

	buffer := make([]byte, fileEntry.EndOfFile)

	for i := 0; i < len(blockList); i++ {
		block := ReadBlock(file, blockList[i])
		for j := 0; j < 512 && i*512+j < fileEntry.EndOfFile; j++ {
			buffer[i*512+j] = block[j]
		}
	}

	return buffer, nil
}

func WriteFile(file *os.File, path string, fileType int, auxType int, buffer []byte) error {
	directory, fileName := GetDirectoryAndFileNameFromPath(path)

	DeleteFile(file, path)

	// get list of blocks to write file to
	blockList := createBlockList(file, len(buffer))

	fileEntry, err := GetFreeFileEntryInDirectory(file, directory)
	if err != nil {
		return err
	}

	// seedling file
	if len(buffer) <= 0x200 {
		WriteBlock(file, blockList[0], buffer)
		fileEntry.StorageType = StorageSeedling
	}

	// sapling file needs index block
	if len(buffer) > 0x200 && len(buffer) <= 0x20000 {
		fileEntry.StorageType = StorageSapling

		// write index block with pointers to data blocks
		indexBuffer := make([]byte, 512)
		for i := 1; i < 256; i++ {
			if i < len(blockList) {
				indexBuffer[i] = byte(blockList[i] & 0x00FF)
				indexBuffer[i+256] = byte(blockList[i] >> 8)
			}
		}
		fmt.Println("writing index block")
		WriteBlock(file, blockList[0], indexBuffer)

		// write all data blocks
		blockBuffer := make([]byte, 512)
		blockPointer := 0
		blockIndexNumber := 1
		for i := 0; i < len(buffer); i++ {
			blockBuffer[blockPointer] = buffer[i]
			if blockPointer == 511 {
				fmt.Printf("A i: %d, blockIndexNumber: %d, blockPointer: %d blockList[blockIndexNumber]: %d\n", i, blockIndexNumber, blockPointer, blockList[blockIndexNumber])
				fmt.Println(blockIndexNumber)
				WriteBlock(file, blockList[blockIndexNumber], blockBuffer)
				blockPointer = 0
				blockIndexNumber++
			} else if i == len(buffer)-1 {
				fmt.Printf("B i: %d, blockIndexNumber: %d, blockPointer: %d\n", i, blockIndexNumber, blockPointer)
				for j := blockPointer; j < 512; j++ {
					blockBuffer[j] = 0
				}
				WriteBlock(file, blockList[blockIndexNumber], blockBuffer)
			} else {
				blockPointer++
			}
		}
	}

	// TODO: add tree file

	// update volume bitmap
	volumeBitmap := ReadVolumeBitmap(file)
	for i := 0; i < len(blockList); i++ {
		markBlockInVolumeBitmap(volumeBitmap, blockList[i])
	}
	writeVolumeBitmap(file, volumeBitmap)

	// add file entry to directory
	fileEntry.FileName = fileName
	fileEntry.BlocksUsed = len(blockList)
	fileEntry.CreationTime = time.Now()
	fileEntry.ModifiedTime = time.Now()
	fileEntry.AuxType = auxType
	fileEntry.EndOfFile = len(buffer)
	fileEntry.FileType = fileType
	fileEntry.KeyPointer = blockList[0]
	fileEntry.Access = 0b11100011

	writeFileEntry(file, fileEntry)

	return nil
}

func getBlocklist(file *os.File, fileEntry FileEntry) []int {
	blocks := make([]int, fileEntry.BlocksUsed)

	switch fileEntry.StorageType {
	case StorageSeedling:
		blocks[0] = fileEntry.KeyPointer
	case StorageSapling:
		index := ReadBlock(file, fileEntry.KeyPointer)
		blocks[0] = fileEntry.KeyPointer
		for i := 1; i < fileEntry.BlocksUsed-1; i++ {
			blocks[i] = int(index[i]) + int(index[i+256])*256
		}
	case StorageTree:
		masterIndex := ReadBlock(file, fileEntry.KeyPointer)
		blocks[0] = fileEntry.KeyPointer
		for i := 0; i < 128; i++ {
			index := ReadBlock(file, int(masterIndex[i])+int(masterIndex[i+256])*256)
			for j := 0; j < 256 && i*256+j < fileEntry.BlocksUsed; j++ {
				blocks[i*256+j] = int(index[j]) + int(index[j+256])*256
			}
		}
	}

	return blocks
}

func createBlockList(file *os.File, fileSize int) []int {
	numberOfBlocks := fileSize / 512
	if fileSize%512 > 0 {
		numberOfBlocks++
	}
	if fileSize > 0x200 && fileSize <= 0x20000 {
		numberOfBlocks++ // add index block
	}
	if fileSize > 0x20000 {
		// add master index block
		numberOfBlocks++
		// add index blocks for each 128 blocks
		numberOfBlocks += numberOfBlocks / 128
		// add index block for any remaining blocks
		if numberOfBlocks%128 > 0 {
			numberOfBlocks++
		}
	}
	volumeBitmap := ReadVolumeBitmap(file)
	blockList := findFreeBlocks(volumeBitmap, numberOfBlocks)

	return blockList
}

func GetFileEntry(file *os.File, path string) (FileEntry, error) {
	directory, fileName := GetDirectoryAndFileNameFromPath(path)

	_, _, fileEntries := ReadDirectory(file, directory)

	if fileEntries == nil {
		return FileEntry{}, errors.New("File entry not found")
	}

	var fileEntry FileEntry

	for i := 0; i < len(fileEntries); i++ {
		if fileEntries[i].FileName == fileName {
			fileEntry = fileEntries[i]
		}
	}

	return fileEntry, nil
}

func GetDirectoryAndFileNameFromPath(path string) (string, string) {
	path = strings.ToUpper(path)
	paths := strings.Split(path, "/")

	var directoryBuilder strings.Builder

	for i := 1; i < len(paths)-1; i++ {
		directoryBuilder.WriteString("/")
		directoryBuilder.WriteString(paths[i])
	}

	directory := directoryBuilder.String()
	fileName := paths[len(paths)-1]

	return directory, fileName
}

func DeleteFile(file *os.File, path string) error {
	fileEntry, err := GetFileEntry(file, path)
	if err != nil {
		return errors.New("File not found")
	}
	if fileEntry.StorageType == StorageDeleted {
		return errors.New("File already deleted")
	}

	// free the blocks
	blocks := getBlocklist(file, fileEntry)
	volumeBitmap := ReadVolumeBitmap(file)
	for i := 0; i < len(blocks); i++ {
		freeBlockInVolumeBitmap(volumeBitmap, blocks[i])
	}
	writeVolumeBitmap(file, volumeBitmap)

	// zero out directory entry
	fileEntry.StorageType = 0
	fileEntry.FileName = ""
	writeFileEntry(file, fileEntry)

	// decrement the directory entry count
	directoryBlock := ReadBlock(file, fileEntry.HeaderPointer)
	directoryHeader := parseDirectoryHeader(directoryBlock, fileEntry.HeaderPointer)

	directoryHeader.ActiveFileCount--
	writeDirectoryHeader(file, directoryHeader, fileEntry.HeaderPointer)

	return nil
}
