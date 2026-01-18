// Copyright Terence J. Boldt (c)2021-2024
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
	var block []byte

	for i := uint32(0); i < uint32(len(blockList)); i++ {
		if blockList[i] == 0 {
			block = zeroData()
		} else {
			block, err = ReadBlock(reader, blockList[i])
			if err != nil {
				return nil, err
			}
		}
		for j := uint32(0); j < 512 && uint32(i)*512+uint32(j) < fileEntry.EndOfFile; j++ {
			buffer[i*512+j] = block[j]
		}
	}

	return buffer, nil
}

// WriteFile writes a file to a ProDOS volume from a byte array
func WriteFile(readerWriter ReaderWriterAt, path string, fileType uint8, auxType uint16, createdTime time.Time, modifiedTime time.Time, buffer []byte) error {
	directory, fileName := GetDirectoryAndFileNameFromPath(path)

	if len(fileName) > 15 {
		return errors.New("filename too long")
	}

	existingFileEntry, _ := GetFileEntry(readerWriter, path)
	if existingFileEntry.StorageType != StorageDeleted {
		return errors.New(("file already exists"))
	}

	// get list of blocks to write file to
	blockList, err := createBlockList(readerWriter, uint32(len(buffer)))
	if err != nil {
		return err
	}

	// seedling file
	if len(buffer) <= 0x200 {
		writeSeedlingFile(readerWriter, buffer, blockList)
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
	fileEntry.BlocksUsed = uint16(len(blockList))
	fileEntry.CreationTime = createdTime
	fileEntry.ModifiedTime = modifiedTime
	fileEntry.AuxType = auxType
	fileEntry.EndOfFile = uint32(len(buffer))
	fileEntry.FileType = fileType
	fileEntry.KeyPointer = blockList[0]
	fileEntry.Version = 0x24
	fileEntry.MinVersion = 0x00
	fileEntry.Access = 0b11100011
	if len(blockList) == 1 {
		fileEntry.StorageType = StorageSeedling
	} else if len(blockList) <= 257 {
		fileEntry.StorageType = StorageSapling
	} else {
		fileEntry.StorageType = StorageTree
	}

	err = writeFileEntry(readerWriter, fileEntry)
	if err != nil {
		return err
	}
	return incrementFileCount(readerWriter, fileEntry)
}

func zeroData() []byte {
	return make([]byte, 512)
}

func incrementFileCount(readerWriter ReaderWriterAt, fileEntry FileEntry) error {
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
	// DumpFileEntry(fileEntry)
	// oldDirectoryBlock, _ := ReadBlock(readerWriter, fileEntry.DirectoryBlock)
	// DumpBlock(oldDirectoryBlock)

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
		if blocks[i] != 0 {
			freeBlockInVolumeBitmap(volumeBitmap, blocks[i])
		}
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
	err = writeFileEntry(readerWriter, fileEntry)
	if err != nil {
		return err
	}

	return nil
}

// FileExists return true if the file exists
func FileExists(reader io.ReaderAt, path string) (bool, error) {
	fileEntry, _ := GetFileEntry(reader, path)
	return fileEntry.StorageType != StorageDeleted, nil
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

func updateVolumeBitmap(readerWriter ReaderWriterAt, blockList []uint16) error {

	volumeBitmap, err := ReadVolumeBitmap(readerWriter)
	if err != nil {
		fmt.Printf("%s", err)
		return err
	}
	for i := uint16(0); i < uint16(len(blockList)); i++ {
		markBlockInVolumeBitmap(volumeBitmap, blockList[i])
	}
	return writeVolumeBitmap(readerWriter, volumeBitmap)
}

func writeSeedlingFile(writer io.WriterAt, buffer []byte, blockList []uint16) {
	WriteBlock(writer, blockList[0], buffer)
}

func writeSaplingFile(writer io.WriterAt, buffer []byte, blockList []uint16) {
	// write index block with pointers to data blocks
	indexBuffer := make([]byte, 512)
	for i := 0; i < 256; i++ {
		if i < len(blockList)-1 {
			indexBuffer[i] = byte(blockList[i+1] & 0x00FF)
			indexBuffer[i+256] = byte(blockList[i+1] >> 8)
		} else {
			indexBuffer[i] = 0
			indexBuffer[i+256] = 0
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

func writeTreeFile(writer io.WriterAt, buffer []byte, blockList []uint16) {
	// write master index block with pointers to index blocks
	indexBuffer := make([]byte, 512)
	numberOfIndexBlocks := len(blockList)/256 + 1
	if len(blockList)%256 == 0 {
		numberOfIndexBlocks--
	}
	for i := 0; i < 256; i++ {
		if i < numberOfIndexBlocks {
			indexBuffer[i] = byte(blockList[i+1] & 0x00FF)
			indexBuffer[i+256] = byte(blockList[i+1] >> 8)
		} else {
			indexBuffer[i] = 0
			indexBuffer[i+256] = 0
		}
	}
	WriteBlock(writer, blockList[0], indexBuffer)
	numberOfIndexBlocks++

	// write index blocks
	for i := 0; i < len(blockList)/256+1; i++ {
		for j := 0; j < 256; j++ {
			if i*256+j < len(blockList)-numberOfIndexBlocks {
				indexBuffer[j] = byte(blockList[i*256+numberOfIndexBlocks+j] & 0x00FF)
				indexBuffer[j+256] = byte(blockList[i*256+j+2] >> 8)
			} else {
				indexBuffer[j] = 0
				indexBuffer[j+256] = 0
			}
		}
		WriteBlock(writer, blockList[i+1], indexBuffer)
	}

	// write all data blocks
	blockBuffer := make([]byte, 512)
	blockPointer := 0
	blockIndexNumber := numberOfIndexBlocks
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

func getDataBlocklist(reader io.ReaderAt, fileEntry FileEntry) ([]uint16, error) {
	return getBlocklist(reader, fileEntry, true)
}

func getAllBlockList(reader io.ReaderAt, fileEntry FileEntry) ([]uint16, error) {
	return getBlocklist(reader, fileEntry, false)
}

// Returns all blocks, including index blocks
func getBlocklist(reader io.ReaderAt, fileEntry FileEntry, dataOnly bool) ([]uint16, error) {
	blocks := make([]uint16, fileEntry.BlocksUsed)

	switch fileEntry.StorageType {
	case StorageSeedling:
		blocks[0] = fileEntry.KeyPointer
		return blocks, nil
	case StorageSapling:
		index, err := ReadBlock(reader, fileEntry.KeyPointer)
		if err != nil {
			return nil, err
		}
		blockOffset := uint16(1)
		blocks[0] = fileEntry.KeyPointer
		for i := uint16(0); i < fileEntry.BlocksUsed-1; i++ {
			blocks[i+blockOffset] = uint16(index[i]) + uint16(index[i+256])*256
		}
		if dataOnly {
			return blocks[1:], nil
		}
		return blocks, nil
	case StorageTree:
		// the number of dataBlocks in the list can be longer than the actual blocks used
		// because of sparse files
		dataBlocks := make([]uint16, fileEntry.EndOfFile/512+1)
		// this can be more than needed
		numberOfIndexBlocks := fileEntry.EndOfFile/512 + 2
		indexBlocks := make([]uint16, numberOfIndexBlocks)
		masterIndex, err := ReadBlock(reader, fileEntry.KeyPointer)
		if err != nil {
			return nil, err
		}
		numberOfDataBlocks := 0

		indexBlocks[0] = fileEntry.KeyPointer
		indexBlockCount := 1
		bytesRemaining := fileEntry.EndOfFile

		for i := uint16(0); i < 128; i++ {
			indexBlock := uint16(masterIndex[i]) + uint16(masterIndex[i+256])*256
			// even with sparse files, the index blocks should not be 0, only data blocks can be
			if indexBlock == 0 {
				break
			}
			indexBlocks[indexBlockCount] = indexBlock
			indexBlockCount++
			index, err := ReadBlock(reader, indexBlock)
			if err != nil {
				return nil, err
			}
			for j := uint16(0); j < 256 && bytesRemaining > 0; j++ {
				dataBlocks[numberOfDataBlocks] = uint16(index[j]) + uint16(index[j+256])*256
				numberOfDataBlocks++
				if bytesRemaining < 512 {
					bytesRemaining = 0
				} else {
					bytesRemaining -= 512
				}
			}
		}

		if dataOnly {
			return dataBlocks[0:numberOfDataBlocks], nil
		}

		blocks = append(indexBlocks[0:numberOfIndexBlocks], dataBlocks[0:numberOfDataBlocks]...)
		return blocks, nil
	}

	return nil, errors.New("unsupported file storage type")
}

func createBlockList(reader io.ReaderAt, fileSize uint32) ([]uint16, error) {
	numberOfBlocks := uint16(fileSize / 512)

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

	return blockList[0:numberOfBlocks], nil
}

// GetFileEntry returns a file entry for the given path
func GetFileEntry(reader io.ReaderAt, path string) (FileEntry, error) {
	directory, fileName := GetDirectoryAndFileNameFromPath(path)
	_, _, fileEntries, err := ReadDirectory(reader, directory)
	if err != nil {
		return FileEntry{}, err
	}

	if len(fileEntries) == 0 {
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
