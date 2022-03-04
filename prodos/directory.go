// Copyright Terence J. Boldt (c)2021-2022
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides access to read, write, delete
// fand parse directories on a ProDOS drive image

package prodos

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// VolumeHeader from ProDOS
type VolumeHeader struct {
	VolumeName       string
	CreationTime     time.Time
	ActiveFileCount  int
	BitmapStartBlock int
	TotalBlocks      int
	NextBlock        int
	EntryLength      int
	EntriesPerBlock  int
	MinVersion       int
	Version          int
}

// DirectoryHeader from ProDOS
type DirectoryHeader struct {
	Name            string
	ActiveFileCount int
	StartingBlock   int
	PreviousBlock   int
	NextBlock       int
}

const (
	// StorageDeleted signifies file is deleted
	StorageDeleted = 0
	// StorageSeedling signifies file is <= 512 bytes
	StorageSeedling = 1
	// StorageSapling signifies file is > 512 bytes and <= 128 KB
	StorageSapling = 2
	// StorageTree signifies file is > 128 KB and <= 16 MB
	StorageTree = 3
	// StoragePascal signifies pascal storage area
	StoragePascal = 4
	// StorageDirectory signifies directory
	StorageDirectory = 13
)

// FileEntry from ProDOS
type FileEntry struct {
	StorageType     int
	FileName        string
	FileType        int
	CreationTime    time.Time
	KeyPointer      int
	Version         int
	MinVersion      int
	BlocksUsed      int
	EndOfFile       int
	Access          int
	AuxType         int
	ModifiedTime    time.Time
	HeaderPointer   int
	DirectoryBlock  int
	DirectoryOffset int
}

// ReadDirectory reads the directory information from a specified path
// on a ProDOS image
func ReadDirectory(reader io.ReaderAt, path string) (VolumeHeader, DirectoryHeader, []FileEntry, error) {
	buffer, err := ReadBlock(reader, 2)
	if err != nil {
		return VolumeHeader{}, DirectoryHeader{}, nil, err
	}

	volumeHeader := parseVolumeHeader(buffer)

	if len(path) == 0 {
		path = fmt.Sprintf("/%s", volumeHeader.VolumeName)
	}

	path = strings.ToUpper(path)
	paths := strings.Split(path, "/")

	directoryHeader, fileEntries, err := getFileEntriesInDirectory(reader, 2, 1, paths)
	if err != nil {
		return VolumeHeader{}, DirectoryHeader{}, nil, err
	}

	return volumeHeader, directoryHeader, fileEntries, nil
}

func getFreeFileEntryInDirectory(reader io.ReaderAt, directory string) (FileEntry, error) {
	_, directoryHeader, _, err := ReadDirectory(reader, directory)
	if err != nil {
		return FileEntry{}, err
	}
	//DumpDirectoryHeader(directoryHeader)
	blockNumber := directoryHeader.StartingBlock
	buffer, err := ReadBlock(reader, blockNumber)
	if err != nil {
		return FileEntry{}, err
	}
	entryOffset := 43 // start at offset after header
	entryNumber := 2  // header is essentially the first entry so start at 2

	for {
		if entryNumber > 13 {
			blockNumber = int(buffer[2]) + int(buffer[3])*256
			// if we ran out of blocks in the directory, return empty
			// TODO: expand the directory to add more entries
			if blockNumber == 0 {
				return FileEntry{}, errors.New("No free file entries found")
			}
			// else read the next block in the directory
			buffer, err = ReadBlock(reader, blockNumber)
			if err != nil {
				return FileEntry{}, nil
			}
			entryOffset = 4
			entryNumber = 1
		}
		fileEntry := parseFileEntry(buffer[entryOffset:entryOffset+40], blockNumber, entryOffset)

		if fileEntry.StorageType == StorageDeleted {
			fileEntry.DirectoryBlock = blockNumber
			fileEntry.DirectoryOffset = entryOffset
			fileEntry.HeaderPointer = directoryHeader.StartingBlock
			return fileEntry, nil
		}

		entryNumber++
		entryOffset += 39
	}
}

func getFileEntriesInDirectory(reader io.ReaderAt, blockNumber int, currentPath int, paths []string) (DirectoryHeader, []FileEntry, error) {
	buffer, err := ReadBlock(reader, blockNumber)
	if err != nil {
		return DirectoryHeader{}, nil, err
	}

	directoryHeader := parseDirectoryHeader(buffer, blockNumber)

	fileEntries := make([]FileEntry, directoryHeader.ActiveFileCount)
	entryOffset := 43 // start at offset after header
	activeEntries := 0
	entryNumber := 2 // header is essentially the first entry so start at 2

	nextBlock := directoryHeader.NextBlock

	matchedDirectory := (currentPath == len(paths)-1) && (paths[currentPath] == directoryHeader.Name)

	if !matchedDirectory && (currentPath == len(paths)-1) {
		// path not matched by last path part
		return DirectoryHeader{}, nil, errors.New("path not matched")
	}

	for {
		if entryNumber > 13 {
			entryOffset = 4
			entryNumber = 1
			if blockNumber == 0 {
				return DirectoryHeader{}, nil, nil
			}
			buffer, err = ReadBlock(reader, nextBlock)
			if err != nil {
				return DirectoryHeader{}, nil, err
			}
			nextBlock = int(buffer[2]) + int(buffer[3])*256
		}
		fileEntry := parseFileEntry(buffer[entryOffset:entryOffset+40], blockNumber, entryOffset)

		if fileEntry.StorageType != StorageDeleted {
			if matchedDirectory && activeEntries == directoryHeader.ActiveFileCount {
				return directoryHeader, fileEntries[0:activeEntries], nil
			}
			if matchedDirectory {
				fileEntries[activeEntries] = fileEntry
			} else if !matchedDirectory && fileEntry.FileType == 15 && paths[currentPath+1] == fileEntry.FileName {
				return getFileEntriesInDirectory(reader, fileEntry.KeyPointer, currentPath+1, paths)
			}
			activeEntries++
		}

		entryNumber++
		entryOffset += 39
	}
}

func parseFileEntry(buffer []byte, blockNumber int, entryOffset int) FileEntry {
	storageType := int(buffer[0] >> 4)
	fileNameLength := int(buffer[0] & 15)
	fileName := string(buffer[1 : fileNameLength+1])
	fileType := int(buffer[16])
	startingBlock := int(buffer[17]) + int(buffer[18])*256
	blocksUsed := int(buffer[19]) + int(buffer[20])*256
	endOfFile := int(buffer[21]) + int(buffer[22])*256 + int(buffer[23])*65536
	creationTime := DateTimeFromProDOS(buffer[24:28])
	version := int(buffer[28])
	minVersion := int(buffer[29])
	access := int(buffer[30])
	auxType := int(buffer[31]) + int(buffer[32])*256
	modifiedTime := DateTimeFromProDOS((buffer[33:37]))
	headerPointer := int(buffer[0x25]) + int(buffer[0x26])*256

	fileEntry := FileEntry{
		StorageType:     storageType,
		FileName:        fileName,
		FileType:        fileType,
		CreationTime:    creationTime,
		Version:         version,
		MinVersion:      minVersion,
		KeyPointer:      startingBlock,
		BlocksUsed:      blocksUsed,
		EndOfFile:       endOfFile,
		Access:          access,
		AuxType:         auxType,
		ModifiedTime:    modifiedTime,
		HeaderPointer:   headerPointer,
		DirectoryBlock:  blockNumber,
		DirectoryOffset: entryOffset,
	}

	return fileEntry
}

func writeFileEntry(writer io.WriterAt, fileEntry FileEntry) {
	buffer := make([]byte, 39)
	buffer[0] = byte(fileEntry.StorageType)<<4 + byte(len(fileEntry.FileName))
	for i := 0; i < len(fileEntry.FileName); i++ {
		buffer[i+1] = fileEntry.FileName[i]
	}
	buffer[0x10] = byte(fileEntry.FileType)
	buffer[0x11] = byte(fileEntry.KeyPointer & 0xFF)
	buffer[0x12] = byte(fileEntry.KeyPointer >> 8)
	buffer[0x13] = byte(fileEntry.BlocksUsed & 0xFF)
	buffer[0x14] = byte(fileEntry.BlocksUsed >> 8)
	buffer[0x15] = byte(fileEntry.EndOfFile & 0x0000FF)
	buffer[0x16] = byte(fileEntry.EndOfFile & 0x00FF00 >> 8)
	buffer[0x17] = byte(fileEntry.EndOfFile & 0xFF0000 >> 16)
	creationTime := DateTimeToProDOS(fileEntry.CreationTime)
	for i := 0; i < 4; i++ {
		buffer[0x18+i] = creationTime[i]
	}
	buffer[0x1C] = byte(fileEntry.Version)
	buffer[0x1D] = byte(fileEntry.MinVersion)
	buffer[0x1E] = byte(fileEntry.Access)
	buffer[0x1F] = byte(fileEntry.AuxType & 0x00FF)
	buffer[0x20] = byte(fileEntry.AuxType >> 8)
	modifiedTime := DateTimeToProDOS(fileEntry.CreationTime)
	for i := 0; i < 4; i++ {
		buffer[0x21+i] = modifiedTime[i]
	}
	buffer[0x25] = byte(fileEntry.HeaderPointer & 0x00FF)
	buffer[0x26] = byte(fileEntry.HeaderPointer >> 8)

	_, err := writer.WriteAt(buffer, int64(fileEntry.DirectoryBlock*512+fileEntry.DirectoryOffset))
	if err != nil {

	}
}

func parseVolumeHeader(buffer []byte) VolumeHeader {
	nextBlock := int(buffer[2]) + int(buffer[3])*256
	filenameLength := buffer[4] & 15
	volumeName := string(buffer[5 : filenameLength+5])
	creationTime := DateTimeFromProDOS(buffer[28:32])
	version := int(buffer[32])
	minVersion := int(buffer[33])
	entryLength := int(buffer[35])
	entriesPerBlock := int(buffer[36])
	fileCount := int(buffer[37]) + int(buffer[38])*256
	bitmapBlock := int(buffer[39]) + int(buffer[40])*256
	totalBlocks := int(buffer[41]) + int(buffer[42])*256

	if version > 0 || minVersion > 0 {
		panic("Unsupported ProDOS version")
	}

	volumeHeader := VolumeHeader{
		VolumeName:       volumeName,
		CreationTime:     creationTime,
		ActiveFileCount:  fileCount,
		BitmapStartBlock: bitmapBlock,
		TotalBlocks:      totalBlocks,
		NextBlock:        nextBlock,
		EntriesPerBlock:  entriesPerBlock,
		EntryLength:      entryLength,
		MinVersion:       minVersion,
		Version:          version,
	}
	return volumeHeader
}

func parseDirectoryHeader(buffer []byte, blockNumber int) DirectoryHeader {
	previousBlock := int(buffer[0x00]) + int(buffer[0x01])*256
	nextBlock := int(buffer[0x02]) + int(buffer[0x03])*256
	filenameLength := buffer[0x04] & 15
	name := string(buffer[0x05 : filenameLength+0x05])
	fileCount := int(buffer[0x25]) + int(buffer[0x26])*256

	directoryEntry := DirectoryHeader{
		PreviousBlock:   previousBlock,
		NextBlock:       nextBlock,
		StartingBlock:   blockNumber,
		Name:            name,
		ActiveFileCount: fileCount,
	}

	return directoryEntry
}

func writeDirectoryHeader(readerWriter ReaderWriterAt, directoryHeader DirectoryHeader) error {
	buffer, err := ReadBlock(readerWriter, directoryHeader.StartingBlock)
	if err != nil {
		return err
	}
	buffer[0x00] = byte(directoryHeader.PreviousBlock & 0x00FF)
	buffer[0x01] = byte(directoryHeader.PreviousBlock >> 8)
	buffer[0x02] = byte(directoryHeader.NextBlock & 0x00FF)
	buffer[0x03] = byte(directoryHeader.NextBlock >> 8)
	buffer[0x04] = buffer[0x04] | byte(len(directoryHeader.Name))
	for i := 0; i < len(directoryHeader.Name); i++ {
		buffer[0x05+i] = directoryHeader.Name[i]
	}
	buffer[0x25] = byte(directoryHeader.ActiveFileCount & 0x00FF)
	buffer[0x26] = byte(directoryHeader.ActiveFileCount >> 8)
	WriteBlock(readerWriter, directoryHeader.StartingBlock, buffer)

	return nil
}
