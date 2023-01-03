// Copyright Terence J. Boldt (c)2021-2022
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides access to read, write and delete
// files on a ProDOS drive image

package prodos

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// LoadFile loads in a file from a ProDOS volume into a byte array
func LoadFile(reader io.ReaderAt, path string) ([]byte, error) {
	fileEntry, err := GetFileEntry(reader, path)
	if err != nil {
		return nil, err
	}

	blockList, err := getDataBlocklist(reader, fileEntry)
	if err != nil {
		return nil, err
	}

	buffer := make([]byte, fileEntry.EndOfFile)

	for i := 0; i < len(blockList); i++ {
		block, err := ReadBlock(reader, blockList[i])
		if err != nil {
			return nil, err
		}
		for j := 0; j < 512 && i*512+j < fileEntry.EndOfFile; j++ {
			buffer[i*512+j] = block[j]
		}
	}

	return buffer, nil
}

// WriteFile writes a file to a ProDOS volume from a byte array
func WriteFile(readerWriter ReaderWriterAt, path string, fileType int, auxType int, buffer []byte) error {
	directory, fileName := GetDirectoryAndFileNameFromPath(path)

	existingFileEntry, _ := GetFileEntry(readerWriter, path)
	if existingFileEntry.StorageType != StorageDeleted {
		DeleteFile(readerWriter, path)
	}

	// get list of blocks to write file to
	blockList, err := createBlockList(readerWriter, len(buffer))
	if err != nil {
		return err
	}

	// seedling file
	if len(buffer) <= 0x200 {
		WriteBlock(readerWriter, blockList[0], buffer)
	}

	// sapling file needs index block
	if len(buffer) > 0x200 && len(buffer) <= 0x20000 {
		writeSaplingFile(readerWriter, buffer, blockList)
	}

	// tree file needs master index and index blocks
	if len(buffer) > 0x20000 && len(buffer) <= 0x1000000 {
		writeTreeFile(readerWriter, buffer, blockList)
	}

	if len(buffer) > 0x1000000 {
		return errors.New("files > 16MB not supported by ProDOS")
	}

	updateVolumeBitmap(readerWriter, blockList)

	// add file entry to directory
	fileEntry, err := getFreeFileEntryInDirectory(readerWriter, directory)
	if err != nil {
		return err
	}
	fileEntry.FileName = fileName
	fileEntry.BlocksUsed = len(blockList)
	fileEntry.CreationTime = time.Now()
	fileEntry.ModifiedTime = time.Now()
	fileEntry.AuxType = auxType
	fileEntry.EndOfFile = len(buffer)
	fileEntry.FileType = fileType
	fileEntry.KeyPointer = blockList[0]
	fileEntry.Access = 0b11100011
	if len(blockList) == 1 {
		fileEntry.StorageType = StorageSeedling
	} else if len(blockList) <= 257 {
		fileEntry.StorageType = StorageSapling
	} else {
		fileEntry.StorageType = StorageTree
	}

	writeFileEntry(readerWriter, fileEntry)

	// increment file count
	directoryHeaderBlock, err := ReadBlock(readerWriter, fileEntry.HeaderPointer)
	if err != nil {
		return err
	}
	directoryHeader := parseDirectoryHeader(directoryHeaderBlock, fileEntry.HeaderPointer)
	directoryHeader.ActiveFileCount++
	writeDirectoryHeader(readerWriter, directoryHeader)

	return nil
}

// DeleteFile deletes a file from a ProDOS volume
func DeleteFile(readerWriter ReaderWriterAt, path string) error {
	fileEntry, err := GetFileEntry(readerWriter, path)
	if err != nil {
		return errors.New("file not found")
	}
	if fileEntry.StorageType == StorageDeleted {
		return errors.New("file already deleted")
	}
	if fileEntry.StorageType == StorageDirectory {
		return errors.New("directory deletion not supported")
	}

	// free the blocks
	blocks, err := getAllBlockList(readerWriter, fileEntry)
	if err != nil {
		return err
	}
	volumeBitmap, err := ReadVolumeBitmap(readerWriter)
	if err != nil {
		return err
	}
	for i := 0; i < len(blocks); i++ {
		freeBlockInVolumeBitmap(volumeBitmap, blocks[i])
	}
	writeVolumeBitmap(readerWriter, volumeBitmap)

	// decrement the directory entry count
	directoryBlock, err := ReadBlock(readerWriter, fileEntry.HeaderPointer)
	if err != nil {
		return err
	}
	directoryHeader := parseDirectoryHeader(directoryBlock, fileEntry.HeaderPointer)

	directoryHeader.ActiveFileCount--
	writeDirectoryHeader(readerWriter, directoryHeader)

	// zero out directory entry
	fileEntry.StorageType = 0
	fileEntry.FileName = ""
	writeFileEntry(readerWriter, fileEntry)

	return nil
}

// GetDirectoryAndFileNameFromPath gets the directory and filename from a path
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

func updateVolumeBitmap(readerWriter ReaderWriterAt, blockList []int) error {

	volumeBitmap, err := ReadVolumeBitmap(readerWriter)
	if err != nil {
		fmt.Printf("%s", err)
		return err
	}
	for i := 0; i < len(blockList); i++ {
		markBlockInVolumeBitmap(volumeBitmap, blockList[i])
	}
	return writeVolumeBitmap(readerWriter, volumeBitmap)
}

func writeSaplingFile(writer io.WriterAt, buffer []byte, blockList []int) {
	// write index block with pointers to data blocks
	indexBuffer := make([]byte, 512)
	for i := 0; i < 256; i++ {
		if i < len(blockList)-1 {
			indexBuffer[i] = byte(blockList[i+1] & 0x00FF)
			indexBuffer[i+256] = byte(blockList[i+1] >> 8)
		}
	}
	WriteBlock(writer, blockList[0], indexBuffer)

	// write all data blocks
	blockBuffer := make([]byte, 512)
	blockPointer := 0
	blockIndexNumber := 1
	for i := 0; i < len(buffer); i++ {
		blockBuffer[blockPointer] = buffer[i]
		if blockPointer == 511 {
			WriteBlock(writer, blockList[blockIndexNumber], blockBuffer)
			blockPointer = 0
			blockIndexNumber++
		} else if i == len(buffer)-1 {
			for j := blockPointer; j < 512; j++ {
				blockBuffer[j] = 0
			}
			WriteBlock(writer, blockList[blockIndexNumber], blockBuffer)
		} else {
			blockPointer++
		}
	}
}

func writeTreeFile(writer io.WriterAt, buffer []byte, blockList []int) {
}

func getDataBlocklist(reader io.ReaderAt, fileEntry FileEntry) ([]int, error) {
	return getBlocklist(reader, fileEntry, true)
}

func getAllBlockList(reader io.ReaderAt, fileEntry FileEntry) ([]int, error) {
	return getBlocklist(reader, fileEntry, false)
}

// Returns all blocks, including index blocks
func getBlocklist(reader io.ReaderAt, fileEntry FileEntry, dataOnly bool) ([]int, error) {
	blocks := make([]int, fileEntry.BlocksUsed)

	switch fileEntry.StorageType {
	case StorageSeedling:
		blocks[0] = fileEntry.KeyPointer
		return blocks, nil
	case StorageSapling:
		index, err := ReadBlock(reader, fileEntry.KeyPointer)
		if err != nil {
			return nil, err
		}
		blockOffset := 0
		if !dataOnly {
			blocks[0] = fileEntry.KeyPointer
			blockOffset = 1
		}
		for i := 0; i < fileEntry.BlocksUsed-1; i++ {
			blocks[i+blockOffset] = int(index[i]) + int(index[i+256])*256
		}
		return blocks, nil
	case StorageTree:
		dataBlocks := make([]int, fileEntry.BlocksUsed)
		masterIndex, err := ReadBlock(reader, fileEntry.KeyPointer)
		if err != nil {
			return nil, err
		}
		blockOffset := 0
		if !dataOnly {
			blocks[0] = fileEntry.KeyPointer
			blockOffset = 1
		}
		for i := 0; i < 128; i++ {
			indexBlock := int(masterIndex[i]) + int(masterIndex[i+256])*256
			if indexBlock == 0 {
				break
			}
			if !dataOnly {
				blockOffset++
			}
			index, err := ReadBlock(reader, indexBlock)
			if err != nil {
				return nil, err
			}
			for j := 0; j < 256 && i*256+j < fileEntry.BlocksUsed; j++ {
				if (int(index[j]) + int(index[j+256])*256) == 0 {
					break
				}
				dataBlocks[i*256+j] = int(index[j]) + int(index[j+256])*256
			}
		}

		if dataOnly {
			return dataBlocks, nil
		}

		blocks = append(blocks[blockOffset:], dataBlocks...)
		return blocks, nil
	}

	return nil, errors.New("unsupported file storage type")
}

func createBlockList(reader io.ReaderAt, fileSize int) ([]int, error) {
	numberOfBlocks := fileSize / 512

	if fileSize%512 > 0 {
		numberOfBlocks++
	}

	if fileSize > 0x200 && fileSize <= 0x20000 {
		numberOfBlocks++ // add index block
	}

	if fileSize > 0x20000 && fileSize <= 0x1000000 {
		// add index blocks for each 256 blocks
		numberOfBlocks += numberOfBlocks / 256
		// add index block for any remaining blocks
		if numberOfBlocks%256 > 0 {
			numberOfBlocks++
		}
		// add master index block
		numberOfBlocks++
	}
	if fileSize > 0x1000000 {
		return nil, errors.New("file size too large")
	}

	volumeBitmap, err := ReadVolumeBitmap(reader)
	if err != nil {
		return nil, err
	}

	blockList := findFreeBlocks(volumeBitmap, numberOfBlocks)

	return blockList, nil
}

// GetFileEntry returns a file entry for the given path
func GetFileEntry(reader io.ReaderAt, path string) (FileEntry, error) {
	directory, fileName := GetDirectoryAndFileNameFromPath(path)
	_, _, fileEntries, err := ReadDirectory(reader, directory)
	if err != nil {
		return FileEntry{}, err
	}

	if fileEntries == nil || len(fileEntries) == 0 {
		return FileEntry{}, errors.New("file entry not found")
	}

	var fileEntry FileEntry

	for i := 0; i < len(fileEntries); i++ {
		if fileEntries[i].FileName == fileName {
			fileEntry = fileEntries[i]
		}
	}

	if fileEntry.StorageType == StorageDeleted {
		return FileEntry{}, errors.New("file not found")
	}

	return fileEntry, nil
}
