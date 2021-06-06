package prodos

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type VolumeHeader struct {
	VolumeName       string
	CreationTime     time.Time
	ActiveFileCount  int
	BitmapStartBlock int
	TotalBlocks      int
	NextBlock        int
	EntryLength      int
	EntriesPerBlock  int
}

type DirectoryHeader struct {
	Name            string
	ActiveFileCount int
	NextBlock       int
}

const (
	StorageDeleted   = 0
	StorageSeedling  = 1
	StorageSapling   = 2
	StorageTree      = 3
	StoragePascal    = 4
	StorageDirectory = 13
)

type FileEntry struct {
	StorageType   int
	FileName      string
	FileType      int
	CreationTime  time.Time
	StartingBlock int
	BlocksUsed    int
	EndOfFile     int
	Access        int
	AuxType       int
	ModifiedTime  time.Time
}

func ReadDirectory(driveFileName string, path string) (VolumeHeader, []FileEntry) {
	file, err := os.OpenFile(driveFileName, os.O_RDWR, 0755)
	if err != nil {
		return VolumeHeader{}, nil
	}

	buffer := ReadBlock(file, 2)

	volumeHeader := ParseVolumeHeader(buffer)
	//dumpVolumeHeader(volumeHeader)

	if len(path) == 0 {
		path = fmt.Sprintf("/%s", volumeHeader.VolumeName)
	}

	path = strings.ToUpper(path)
	paths := strings.Split(path, "/")

	fileEntries := getFileEntriesInDirectory(file, 2, 1, paths)

	return volumeHeader, fileEntries
}

func getFileEntriesInDirectory(file *os.File, blockNumber int, currentPath int, paths []string) []FileEntry {
	//fmt.Printf("Parsing '%s'...\n", paths[currentPath])

	buffer := ReadBlock(file, blockNumber)

	directoryHeader := ParseDirectoryHeader(buffer)

	fileEntries := make([]FileEntry, directoryHeader.ActiveFileCount)
	entryOffset := 43 // start at offset after header
	activeEntries := 0
	entryNumber := 2 // header is essentially the first entry so start at 2

	nextBlock := directoryHeader.NextBlock

	matchedDirectory := (currentPath == len(paths)-1) && (paths[currentPath] == directoryHeader.Name)

	if !matchedDirectory && (currentPath == len(paths)-1) {
		// path not matched by last path part
		return nil
	}

	for {
		if entryNumber > 13 {
			entryOffset = 4
			entryNumber = 1
			if blockNumber == 0 {
				return nil
			}
			buffer = ReadBlock(file, nextBlock)
			nextBlock = int(buffer[2]) + int(buffer[3])*256
		}
		fileEntry := parseFileEntry(buffer[entryOffset : entryOffset+40])
		//DumpFileEntry(fileEntry)

		if fileEntry.StorageType != StorageDeleted {
			if matchedDirectory {
				fileEntries[activeEntries] = fileEntry
			} else if !matchedDirectory && fileEntry.FileType == 15 && paths[currentPath+1] == fileEntry.FileName {
				return getFileEntriesInDirectory(file, fileEntry.StartingBlock, currentPath+1, paths)
			}
			activeEntries++
			if matchedDirectory && activeEntries == directoryHeader.ActiveFileCount {
				return fileEntries[0:activeEntries]
			}
		}

		entryNumber++
		entryOffset += 39
	}
}

func parseFileEntry(buffer []byte) FileEntry {
	storageType := int(buffer[0] >> 4)
	fileNameLength := int(buffer[0] & 15)
	fileName := string(buffer[1 : fileNameLength+1])
	fileType := int(buffer[16])
	startingBlock := int(buffer[17]) + int(buffer[18])*256
	blocksUsed := int(buffer[19]) + int(buffer[20])*256
	endOfFile := int(buffer[21]) + int(buffer[22])*256 + int(buffer[23])*65536
	creationTime := DateTimeFromProDOS(buffer[24:28])
	access := int(buffer[30])
	auxType := int(buffer[31]) + int(buffer[32])*256
	modifiedTime := DateTimeFromProDOS((buffer[33:37]))

	fileEntry := FileEntry{
		StorageType:   storageType,
		FileName:      fileName,
		FileType:      fileType,
		CreationTime:  creationTime,
		StartingBlock: startingBlock,
		BlocksUsed:    blocksUsed,
		EndOfFile:     endOfFile,
		Access:        access,
		AuxType:       auxType,
		ModifiedTime:  modifiedTime,
	}

	return fileEntry
}

func ParseVolumeHeader(buffer []byte) VolumeHeader {
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
	}
	return volumeHeader
}

func ParseDirectoryHeader(buffer []byte) DirectoryHeader {
	nextBlock := int(buffer[2]) + int(buffer[3])*256
	filenameLength := buffer[4] & 15
	name := string(buffer[5 : filenameLength+5])
	fileCount := int(buffer[37]) + int(buffer[38])*256

	directoryEntry := DirectoryHeader{
		NextBlock:       nextBlock,
		Name:            name,
		ActiveFileCount: fileCount,
	}

	return directoryEntry
}
