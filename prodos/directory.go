// Copyright Terence J. Boldt (c)2021-2024
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
	ActiveFileCount  uint16
	BitmapStartBlock uint16
	TotalBlocks      uint16
	NextBlock        uint16
	EntryLength      uint8
	EntriesPerBlock  uint8
	MinVersion       uint8
	Version          uint8
}

// DirectoryHeader from ProDOS
type DirectoryHeader struct {
	PreviousBlock     uint16
	NextBlock         uint16
	IsSubDirectory    bool
	Name              string
	CreationTime      time.Time
	Version           uint8
	MinVersion        uint8
	Access            uint8
	EntryLength       uint8
	EntriesPerBlock   uint8
	ActiveFileCount   uint16
	StartingBlock     uint16
	ParentBlock       uint16
	ParentEntry       uint16
	ParentEntryLength uint8
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
	StorageType   uint8
	FileName      string
	FileType      uint8
	CreationTime  time.Time
	KeyPointer    uint16
	Version       uint8
	MinVersion    uint8
	BlocksUsed    uint16
	EndOfFile     uint32
	Access        uint8
	AuxType       uint16
	ModifiedTime  time.Time
	HeaderPointer uint16

	DirectoryBlock  uint16
	DirectoryOffset uint16
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

	// add volume name if not full path
	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s/%s", volumeHeader.VolumeName, path)
	}

	path = strings.ToUpper(path)
	paths := strings.Split(path, "/")

	directoryHeader, fileEntries, err := getFileEntriesInDirectory(reader, 2, 1, paths)
	if err != nil {
		return VolumeHeader{}, DirectoryHeader{}, nil, err
	}

	return volumeHeader, directoryHeader, fileEntries, nil
}

// CreateDirectory creates a directory information of a specified path
// on a ProDOS image
func CreateDirectory(readerWriter ReaderWriterAt, path string) error {
	if len(path) == 0 {
		return errors.New("cannot create directory with path")
	}

	// add volume name if not full path
	path, err := makeFullPath(path, readerWriter)
	if err != nil {
		return err
	}

	parentPath, newDirectory := GetDirectoryAndFileNameFromPath(path)

	existingFileEntry, _ := GetFileEntry(readerWriter, path)
	if existingFileEntry.StorageType != StorageDeleted {
		return errors.New("directory already exists")
	}

	fileEntry, err := getFreeFileEntryInDirectory(readerWriter, parentPath)
	if err != nil {
		errString := fmt.Sprintf("failed to create directory: %s", err)
		return errors.New(errString)
	}

	// get list of blocks to write file to
	blockList, err := createBlockList(readerWriter, 512)
	if err != nil {
		errString := fmt.Sprintf("failed to create directory: %s", err)
		return errors.New(errString)
	}

	updateVolumeBitmap(readerWriter, blockList)

	fileEntry.FileName = newDirectory
	fileEntry.BlocksUsed = 1
	fileEntry.CreationTime = time.Now()
	fileEntry.ModifiedTime = time.Now()
	fileEntry.AuxType = 0
	fileEntry.EndOfFile = 0x200
	fileEntry.FileType = 0x0F
	fileEntry.KeyPointer = blockList[0]
	fileEntry.Access = 0b11100011
	fileEntry.StorageType = StorageDirectory
	fileEntry.Version = 0x24
	fileEntry.MinVersion = 0x00

	err = writeFileEntry(readerWriter, fileEntry)
	if err != nil {
		return err
	}

	err = incrementFileCount(readerWriter, fileEntry)
	if err != nil {
		errString := fmt.Sprintf("failed to create directory: %s", err)
		return errors.New(errString)
	}

	directoryEntry := DirectoryHeader{
		PreviousBlock:     0,
		NextBlock:         0,
		IsSubDirectory:    true,
		Name:              newDirectory,
		CreationTime:      time.Now(),
		Version:           0x24,
		MinVersion:        0,
		Access:            0xE3,
		EntryLength:       0x27,
		EntriesPerBlock:   0x0D,
		ActiveFileCount:   0,
		StartingBlock:     blockList[0],
		ParentBlock:       fileEntry.DirectoryBlock,
		ParentEntry:       uint16(fileEntry.DirectoryOffset-0x04) / 0x27,
		ParentEntryLength: 0x27,
	}

	err = writeDirectoryHeader(readerWriter, directoryEntry)
	if err != nil {
		errString := fmt.Sprintf("failed to create directory: %s", err)
		return errors.New(errString)
	}

	return nil
}

func makeFullPath(path string, reader io.ReaderAt) (string, error) {
	if !strings.HasPrefix(path, "/") {
		buffer, err := ReadBlock(reader, 0x0002)
		if err != nil {
			return "", err
		}

		volumeHeader := parseVolumeHeader(buffer)
		path = fmt.Sprintf("/%s/%s", volumeHeader.VolumeName, path)
	}
	return path, nil
}

func getFreeFileEntryInDirectory(readerWriter ReaderWriterAt, directory string) (FileEntry, error) {
	_, directoryHeader, _, err := ReadDirectory(readerWriter, directory)
	if err != nil {
		return FileEntry{}, err
	}
	blockNumber := directoryHeader.StartingBlock
	buffer, err := ReadBlock(readerWriter, blockNumber)
	if err != nil {
		return FileEntry{}, err
	}
	entryOffset := uint16(43) // start at offset after header
	entryNumber := 2          // header is essentially the first entry so start at 2

	for {
		if entryNumber > 13 {
			nextBlockNumber := uint16(buffer[2]) + uint16(buffer[3])*256
			// if we ran out of blocks in the directory, expand directory or fail
			if nextBlockNumber == 0 {
				if !directoryHeader.IsSubDirectory {
					return FileEntry{}, errors.New("no free file entries found")
				}
				nextBlockNumber, err = expandDirectory(readerWriter, buffer, blockNumber, directoryHeader)
				if err != nil {
					return FileEntry{}, err
				}
			}
			blockNumber = nextBlockNumber
			// else read the next block in the directory
			buffer, err = ReadBlock(readerWriter, blockNumber)
			if err != nil {
				return FileEntry{}, nil
			}

			entryOffset = 4
			entryNumber = 1
		}
		fileEntry := parseFileEntry(buffer[entryOffset:entryOffset+0x28], blockNumber, entryOffset)

		if fileEntry.StorageType == StorageDeleted {
			fileEntry = FileEntry{}
			fileEntry.DirectoryBlock = blockNumber
			fileEntry.DirectoryOffset = entryOffset
			fileEntry.HeaderPointer = directoryHeader.StartingBlock
			return fileEntry, nil
		}

		entryNumber++
		entryOffset += 39
	}
}

func expandDirectory(readerWriter ReaderWriterAt, buffer []byte, blockNumber uint16, directoryHeader DirectoryHeader) (uint16, error) {
	volumeBitMap, err := ReadVolumeBitmap(readerWriter)
	if err != nil {
		errString := fmt.Sprintf("failed to get volume bitmap to expand directory: %s", err)
		return 0, errors.New(errString)
	}
	blockList := findFreeBlocks(volumeBitMap, 1)
	if len(blockList) != 1 {
		return 0, errors.New("failed to get free block to expand directory")
	}

	nextBlockNumber := blockList[0]
	buffer[0x02] = byte(nextBlockNumber & 0x00FF)
	buffer[0x03] = byte(nextBlockNumber >> 8)
	err = WriteBlock(readerWriter, blockNumber, buffer)
	if err != nil {
		errString := fmt.Sprintf("failed to write block to expand directory: %s", err)
		return 0, errors.New(errString)
	}

	buffer = make([]byte, 0x200)
	buffer[0x00] = byte(blockNumber & 0x00FF)
	buffer[0x01] = byte(blockNumber >> 8)
	err = WriteBlock(readerWriter, nextBlockNumber, buffer)
	if err != nil {
		errString := fmt.Sprintf("failed to write new block to expand directory: %s", err)
		return 0, errors.New(errString)
	}

	updateVolumeBitmap(readerWriter, blockList)

	buffer, err = ReadBlock(readerWriter, directoryHeader.ParentBlock)
	if err != nil {
		errString := fmt.Sprintf("failed to read parent block to expand directory: %s", err)
		return 0, errors.New(errString)
	}
	directoryEntryOffset := directoryHeader.ParentEntry*uint16(directoryHeader.EntryLength) + 0x04
	directoryFileEntry := parseFileEntry(buffer[directoryEntryOffset:directoryEntryOffset+0x28], directoryHeader.ParentBlock, directoryHeader.ParentEntry*uint16(directoryHeader.EntryLength)+0x04)
	directoryFileEntry.BlocksUsed++
	directoryFileEntry.EndOfFile += 0x200
	err = writeFileEntry(readerWriter, directoryFileEntry)
	if err != nil {
		return 0, err
	}

	return nextBlockNumber, nil
}

func getFileEntriesInDirectory(reader io.ReaderAt, blockNumber uint16, currentPath int, paths []string) (DirectoryHeader, []FileEntry, error) {
	buffer, err := ReadBlock(reader, blockNumber)
	if err != nil {
		return DirectoryHeader{}, nil, err
	}

	directoryHeader := parseDirectoryHeader(buffer, blockNumber)

	fileEntries := make([]FileEntry, directoryHeader.ActiveFileCount)
	entryOffset := uint16(43) // start at offset after header
	activeEntries := uint16(0)
	entryNumber := uint8(2) // header is essentially the first entry so start at 2

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
			nextBlock = uint16(buffer[2]) + uint16(buffer[3])*256
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

func parseFileEntry(buffer []byte, blockNumber uint16, entryOffset uint16) FileEntry {
	storageType := buffer[0] >> 4
	fileNameLength := buffer[0] & 15
	fileName := string(buffer[1 : fileNameLength+1])
	fileType := buffer[16]
	startingBlock := uint16(buffer[17]) + uint16(buffer[18])*256
	blocksUsed := uint16(buffer[19]) + uint16(buffer[20])*256
	endOfFile := uint32(buffer[21]) + uint32(buffer[22])*256 + uint32(buffer[23])*65536
	creationTime := DateTimeFromProDOS(buffer[24:28])
	version := buffer[28]
	minVersion := buffer[29]
	access := buffer[30]
	auxType := uint16(buffer[31]) + uint16(buffer[32])*256
	modifiedTime := DateTimeFromProDOS((buffer[33:37]))
	headerPointer := uint16(buffer[0x25]) + uint16(buffer[0x26])*256

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

func writeFileEntry(writer io.WriterAt, fileEntry FileEntry) error {
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
	modifiedTime := DateTimeToProDOS(fileEntry.ModifiedTime)
	for i := 0; i < 4; i++ {
		buffer[0x21+i] = modifiedTime[i]
	}
	buffer[0x25] = byte(fileEntry.HeaderPointer & 0x00FF)
	buffer[0x26] = byte(fileEntry.HeaderPointer >> 8)

	//fmt.Printf("Writing file entry at block: %04X offset: %04X\n", fileEntry.DirectoryBlock, fileEntry.DirectoryOffset)
	_, err := writer.WriteAt(buffer, int64(fileEntry.DirectoryBlock)*512+int64(fileEntry.DirectoryOffset))
	if err != nil {
		return err
	}

	return nil
}

func parseVolumeHeader(buffer []byte) VolumeHeader {
	nextBlock := uint16(buffer[2]) + uint16(buffer[3])*256
	filenameLength := buffer[4] & 15
	volumeName := string(buffer[5 : filenameLength+5])
	creationTime := DateTimeFromProDOS(buffer[28:32])
	version := buffer[32]
	minVersion := buffer[33]
	entryLength := buffer[35]
	entriesPerBlock := buffer[36]
	fileCount := uint16(buffer[37]) + uint16(buffer[38])*256
	bitmapBlock := uint16(buffer[39]) + uint16(buffer[40])*256
	totalBlocks := uint16(buffer[41]) + uint16(buffer[42])*256

	if minVersion > 0 {
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

func parseDirectoryHeader(buffer []byte, blockNumber uint16) DirectoryHeader {
	previousBlock := uint16(buffer[0x00]) + uint16(buffer[0x01])*256
	nextBlock := uint16(buffer[0x02]) + uint16(buffer[0x03])*256
	isSubDirectory := (buffer[0x04] & 0xF0) == 0xE0
	filenameLength := buffer[0x04] & 0x0F
	name := string(buffer[0x05 : filenameLength+0x05])
	creationTime := DateTimeFromProDOS(buffer[0x1C:0x20])
	version := buffer[0x20]
	minVersion := buffer[0x21]
	access := buffer[0x22]
	entryLength := buffer[0x23]
	entriesPerBlock := buffer[0x24]
	fileCount := uint16(buffer[0x25]) + uint16(buffer[0x26])*256
	parentBlock := uint16(buffer[0x27]) + uint16(buffer[0x28])*256
	parentEntry := uint16(buffer[0x29])
	parentEntryLength := buffer[0x2A]

	directoryEntry := DirectoryHeader{
		PreviousBlock:     previousBlock,
		NextBlock:         nextBlock,
		StartingBlock:     blockNumber,
		IsSubDirectory:    isSubDirectory,
		Name:              name,
		CreationTime:      creationTime,
		Version:           version,
		MinVersion:        minVersion,
		Access:            access,
		EntryLength:       entryLength,
		EntriesPerBlock:   entriesPerBlock,
		ActiveFileCount:   fileCount,
		ParentBlock:       parentBlock,
		ParentEntry:       parentEntry,
		ParentEntryLength: parentEntryLength,
	}

	return directoryEntry
}

func writeDirectoryHeader(readerWriter ReaderWriterAt, directoryHeader DirectoryHeader) error {
	// Reading back the block preserves values including reserved fields
	buffer, err := ReadBlock(readerWriter, directoryHeader.StartingBlock)
	if err != nil {
		return err
	}
	buffer[0x00] = byte(directoryHeader.PreviousBlock & 0x00FF)
	buffer[0x01] = byte(directoryHeader.PreviousBlock >> 8)
	buffer[0x02] = byte(directoryHeader.NextBlock & 0x00FF)
	buffer[0x03] = byte(directoryHeader.NextBlock >> 8)
	if directoryHeader.IsSubDirectory {
		buffer[0x04] = 0xE0
	} else {
		buffer[0x04] = 0xF0
	}
	buffer[0x04] = buffer[0x04] | byte(len(directoryHeader.Name))
	for i := 0; i < len(directoryHeader.Name); i++ {
		buffer[0x05+i] = directoryHeader.Name[i]
	}
	creationTime := DateTimeToProDOS(directoryHeader.CreationTime)
	for i := 0; i < 4; i++ {
		buffer[0x1C+i] = creationTime[i]
	}
	// Without these reserved bytes, reading the directory causes I/O ERROR
	buffer[0x14] = 0x75
	buffer[0x15] = byte(directoryHeader.Version)
	buffer[0x16] = byte(directoryHeader.MinVersion)
	buffer[0x17] = 0xC3
	buffer[0x18] = 0x0D
	buffer[0x19] = 0x27
	buffer[0x1A] = 0x00
	buffer[0x1B] = 0x00

	buffer[0x20] = byte(directoryHeader.Version)
	buffer[0x21] = byte(directoryHeader.MinVersion)
	buffer[0x22] = byte(directoryHeader.Access)
	buffer[0x23] = byte(directoryHeader.EntryLength)
	buffer[0x24] = byte(directoryHeader.EntriesPerBlock)
	buffer[0x25] = byte(directoryHeader.ActiveFileCount & 0x00FF)
	buffer[0x26] = byte(directoryHeader.ActiveFileCount >> 8)
	buffer[0x27] = byte(directoryHeader.ParentBlock & 0x00FF)
	buffer[0x28] = byte(directoryHeader.ParentBlock >> 8)
	buffer[0x29] = byte(directoryHeader.ParentEntry)
	buffer[0x2A] = byte(directoryHeader.ParentEntryLength)
	WriteBlock(readerWriter, directoryHeader.StartingBlock, buffer)

	return nil
}
