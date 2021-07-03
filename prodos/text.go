package prodos

import (
	"fmt"
	"strings"
	"time"
)

func TimeToString(printTime time.Time) string {
	return fmt.Sprintf("%04d-%s-%02d %02d:%02d",
		printTime.Year(),
		strings.ToUpper(printTime.Month().String()[0:3]),
		printTime.Day(),
		printTime.Hour(),
		printTime.Minute(),
	)
}

func FileTypeToString(fileType int) string {
	switch fileType {
	case 1:
		return "BAD"
	case 4:
		return "TXT"
	case 6:
		return "BIN"
	case 7:
		return "FNT"
	case 15:
		return "DIR"
	case 252:
		return "BAS"
	case 253:
		return "VAR"
	case 255:
		return "SYS"
	default:
		return fmt.Sprintf("$%02X", fileType)
	}
	/*
		File Type      Preferred Use
		$00            Typeless file (SOS and ProDOS)
		$01            Bad block file
		$02 *          Pascal code file
		$03 *          Pascal text file
		$04            ASCII text file (SOS and ProDOS)
		$05 *          Pascal data file
		$06            General binary file (SOS and ProDOS)
		$07 *          Font file
		$08            Graphics screen file
		$09 *          Business BASIC program file
		$0A *          Business BASIC data file
		$0B *          Word Processor file
		$0C *          SOS system file
		$0D,$0E *      SOS reserved
		$0F            Directory file (SOS and ProDOS)
		$10 *          RPS data file
		$11 *          RPS index file
		$12 *          AppleFile discard file
		$13 *          AppleFile model file
		$14 *          AppleFile report format file
		$15 *          Screen Library file
		$16-$18 *      SOS reserved
		$19            AppleWorks Data Base file
		$1A            AppleWorks Word Processor file
		$1B            AppleWorks Spreadsheet file
		$1C-$EE        Reserved
		$EF            Pascal area
		$F0            ProDOS CI added command file
		$F1-$F8        ProDOS user defined files 1-8
		$F9            ProDOS reserved
		$FA            Integer BASIC program file
		$FB            Integer BASIC variable file
		$FC            Applesoft program file
		$FD            Applesoft variables file
		$FE            Relocatable code file (EDASM)
		$FF            ProDOS system file
	*/
}

func DumpFileEntry(fileEntry FileEntry) {
	fmt.Printf("FileName: %s\n", fileEntry.FileName)
	fmt.Printf("Creation time: %d-%s-%d %02d:%02d\n", fileEntry.CreationTime.Year(), fileEntry.CreationTime.Month(), fileEntry.CreationTime.Day(), fileEntry.CreationTime.Hour(), fileEntry.CreationTime.Minute())
	fmt.Printf("Modified time: %d-%s-%d %02d:%02d\n", fileEntry.ModifiedTime.Year(), fileEntry.ModifiedTime.Month(), fileEntry.ModifiedTime.Day(), fileEntry.ModifiedTime.Hour(), fileEntry.ModifiedTime.Minute())
	fmt.Printf("AuxType: %04X\n", fileEntry.AuxType)
	fmt.Printf("EOF: %06X\n", fileEntry.EndOfFile)
	fmt.Printf("Blocks used: %04X\n", fileEntry.BlocksUsed)
	fmt.Printf("Starting block: %04X\n", fileEntry.KeyPointer)
	fmt.Printf("File type: %02X\n", fileEntry.FileType)
	fmt.Printf("Storage type: %02X\n", fileEntry.StorageType)
	fmt.Printf("Header pointer: %04X\n", fileEntry.HeaderPointer)
	fmt.Printf("\n")
}

func DumpVolumeHeader(volumeHeader VolumeHeader) {
	fmt.Printf("Next block: %d\n", volumeHeader.NextBlock)
	fmt.Printf("Volume name: %s\n", volumeHeader.VolumeName)
	fmt.Printf("Creation time: %d-%s-%d %02d:%02d\n", volumeHeader.CreationTime.Year(), volumeHeader.CreationTime.Month(), volumeHeader.CreationTime.Day(), volumeHeader.CreationTime.Hour(), volumeHeader.CreationTime.Minute())
	fmt.Printf("ProDOS version (should be 0): %d\n", volumeHeader.Version)
	fmt.Printf("ProDOS mininum version (should be 0): %d\n", volumeHeader.MinVersion)
	fmt.Printf("Entry length (should be 39): %d\n", volumeHeader.EntryLength)
	fmt.Printf("Entries per block (should be 13): %d\n", volumeHeader.EntriesPerBlock)
	fmt.Printf("File count: %d\n", volumeHeader.ActiveFileCount)
	fmt.Printf("Bitmap starting block: %d\n", volumeHeader.BitmapStartBlock)
	fmt.Printf("Total blocks: %d\n", volumeHeader.TotalBlocks)
}

func DumpDirectoryHeader(directoryHeader DirectoryHeader) {
	fmt.Printf("Name: %s\n", directoryHeader.Name)
	fmt.Printf("File count: %d\n", directoryHeader.ActiveFileCount)
	fmt.Printf("Starting block: %04X\n", directoryHeader.StartingBlock)
	fmt.Printf("Previous block: %04X\n", directoryHeader.PreviousBlock)
	fmt.Printf("Next block: %04X\n", directoryHeader.NextBlock)
}

func DumpBlock(buffer []byte) {
	for i := 0; i < len(buffer); i += 16 {
		fmt.Printf("%04X: ", i)
		for j := i; j < i+16; j++ {
			fmt.Printf("%02X ", buffer[j])
		}
		for j := i; j < i+16; j++ {
			c := buffer[j] & 127
			if c >= 32 && c < 127 {
				fmt.Printf("%c", c)
			} else {
				fmt.Printf(".")
			}
		}
		fmt.Printf("\n")
	}
}

func DumpDirectory(blocksFree int, totalBlocks int, path string, fileEntries []FileEntry) {
	fmt.Printf("%s\n\n", path)
	fmt.Printf(" NAME           TYPE  BLOCKS  MODIFIED          CREATED            ENDFILE  SUBTYPE\n\n")

	for i := 0; i < len(fileEntries); i++ {
		var zeroTime = time.Time{}
		var modifiedTime, createdTime string
		if fileEntries[i].ModifiedTime == zeroTime {
			modifiedTime = "<NO DATE>        "
		} else {
			modifiedTime = TimeToString(fileEntries[i].ModifiedTime)
		}
		if fileEntries[i].CreationTime == zeroTime {
			createdTime = "<NO DATE>        "
		} else {
			createdTime = TimeToString(fileEntries[i].CreationTime)
		}
		fmt.Printf(" %-15s %s %7d  %s %s %8d %8d\n",
			fileEntries[i].FileName,
			FileTypeToString(fileEntries[i].FileType),
			fileEntries[i].BlocksUsed,
			modifiedTime,
			createdTime,
			fileEntries[i].EndOfFile,
			fileEntries[i].AuxType,
		)
	}
	fmt.Printf("\n")
	fmt.Printf("BLOCKS FREE: %5d    BLOCKS USED: %5d      TOTAL BLOCKS: %5d\n", blocksFree, totalBlocks-blocksFree, totalBlocks)
}
