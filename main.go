// Copyright Terence J. Boldt (c)2021-2022
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides a command line utility to read, write and delete
// files and directories on a ProDOS drive image as well as format
// new volumes

package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tjboldt/ProDOS-Utilities/prodos"
)

const version = "0.3.1"

func main() {
	var fileName string
	var pathName string
	var command string
	var outFileName string
	var inFileName string
	var blockNumber int
	var volumeSize int
	var volumeName string
	var fileType int
	var auxType int
	flag.StringVar(&fileName, "d", "", "A ProDOS format drive image")
	flag.StringVar(&pathName, "p", "", "Path name in ProDOS drive image (default is root of volume)")
	flag.StringVar(&command, "c", "ls", "Command to execute: ls, get, put, rm, mkdir, readblock, writeblock, create")
	flag.StringVar(&outFileName, "o", "", "Name of file to write")
	flag.StringVar(&inFileName, "i", "", "Name of file to read")
	flag.IntVar(&volumeSize, "s", 65535, "Number of blocks to create the volume with (default 65535, 64 to 65535, 0x0040 to 0xFFFF hex input accepted)")
	flag.StringVar(&volumeName, "v", "NO.NAME", "Specifiy a name for the volume from 1 to 15 characters")
	flag.IntVar(&blockNumber, "b", 0, "A block number to read/write from 0 to 65535 (0x0000 to 0xFFFF hex input accepted)")
	flag.IntVar(&fileType, "t", 6, "ProDOS FileType: 0x04 for TXT, 0x06 for BIN, 0xFC for BAS, 0xFF for SYS etc.")
	flag.IntVar(&auxType, "a", 0x2000, "ProDOS AuxType from 0 to 65535 (0x0000 to 0xFFFF hex input accepted)")
	flag.Parse()

	if len(fileName) == 0 {
		printReadme()
		flag.PrintDefaults()
		os.Exit(1)
	}

	switch command {
	case "ls":
		file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
		if err != nil {
			fmt.Printf("Failed to open drive image %s:\n  %s", fileName, err)
			os.Exit(1)
		}
		defer file.Close()
		pathName = strings.ToUpper(pathName)
		volumeHeader, _, fileEntries, err := prodos.ReadDirectory(file, pathName)
		if err != nil {
			fmt.Printf("Error: %s", err)
		}
		if len(pathName) == 0 {
			pathName = "/" + volumeHeader.VolumeName
		}
		volumeBitmap, err := prodos.ReadVolumeBitmap(file)
		if err != nil {
			fmt.Printf("Failed to open drive image %s:\n  %s", fileName, err)
			os.Exit(1)
		}
		freeBlocks := prodos.GetFreeBlockCount(volumeBitmap, volumeHeader.TotalBlocks)
		prodos.DumpDirectory(freeBlocks, volumeHeader.TotalBlocks, pathName, fileEntries)
	case "get":
		file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
		if err != nil {
			fmt.Printf("Failed to open drive image %s:\n  %s", fileName, err)
			os.Exit(1)
		}
		defer file.Close()
		if len(pathName) == 0 {
			fmt.Println("Missing pathname")
			os.Exit(1)
		}
		getFile, err := prodos.LoadFile(file, pathName)
		if err != nil {
			fmt.Printf("Failed to read file %s: %s\n", pathName, err)
			os.Exit(1)
		}
		if len(outFileName) == 0 {
			_, outFileName = prodos.GetDirectoryAndFileNameFromPath(pathName)
		}
		outFile, err := os.Create(outFileName)
		if err != nil {
			fmt.Printf("Failed to create output file %s: %s\n", outFileName, err)
			os.Exit(1)
		}
		if strings.HasSuffix(strings.ToLower(outFileName), ".bas") {
			fmt.Fprintf(outFile, prodos.ConvertBasicToText(getFile))
		} else {
			outFile.Write(getFile)
		}
	case "put":
		file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
		if err != nil {
			fmt.Printf("Failed to open drive image %s:\n  %s", fileName, err)
			os.Exit(1)
		}
		defer file.Close()
		if len(pathName) == 0 {
			fmt.Println("Missing pathname")
			os.Exit(1)
		}
		inFile, err := os.ReadFile(inFileName)
		if err != nil {
			fmt.Printf("Failed to open input file %s: %s", inFileName, err)
			os.Exit(1)
		}
		if strings.HasSuffix(strings.ToLower(inFileName), ".bas") {
			inFile, err = prodos.ConvertTextToBasic(string(inFile))
			fileType = 0xFC
			auxType = 0x0801
		}

		// Check for an AppleSingle file as produced by cc65
		if // Magic number
		binary.BigEndian.Uint32(inFile[0x00:]) == 0x00051600 &&
			// Version number
			binary.BigEndian.Uint32(inFile[0x04:]) == 0x00020000 &&
			// Number of entries
			binary.BigEndian.Uint16(inFile[0x18:]) == 0x0002 &&
			// Data Fork ID
			binary.BigEndian.Uint32(inFile[0x1A:]) == 0x00000001 &&
			// Offset
			binary.BigEndian.Uint32(inFile[0x1E:]) == 0x0000003A &&
			// Length
			binary.BigEndian.Uint32(inFile[0x22:]) == uint32(len(inFile))-0x3A &&
			// ProDOS File Info ID
			binary.BigEndian.Uint32(inFile[0x26:]) == 0x0000000B &&
			// Offset
			binary.BigEndian.Uint32(inFile[0x2A:]) == 0x00000032 &&
			// Length
			binary.BigEndian.Uint32(inFile[0x2E:]) == 0x00000008 {

			fileType = int(binary.BigEndian.Uint16(inFile[0x34:]))
			auxType = int(binary.BigEndian.Uint32(inFile[0x36:]))
			inFile = inFile[0x3A:]
			fmt.Printf("AppleSingle (File type: %02X, AuxType: %04X) detected\n", fileType, auxType)
		}
		err = prodos.WriteFile(file, pathName, fileType, auxType, inFile)
		if err != nil {
			fmt.Printf("Failed to write file %s: %s", pathName, err)
		}
	case "readblock":
		fmt.Printf("Reading block 0x%04X (%d):\n\n", blockNumber, blockNumber)
		file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
		if err != nil {
			fmt.Printf("Failed to open drive image %s:\n  %s", fileName, err)
			os.Exit(1)
		}
		defer file.Close()
		block, err := prodos.ReadBlock(file, blockNumber)
		if err != nil {
			fmt.Printf("Failed to open drive image %s:\n  %s", fileName, err)
			os.Exit(1)
		}
		prodos.DumpBlock(block)
	case "writeblock":
		fmt.Printf("Writing block 0x%04X (%d):\n\n", blockNumber, blockNumber)
		file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
		if err != nil {
			fmt.Printf("Failed to open drive image %s:\n  %s", fileName, err)
			os.Exit(1)
		}
		defer file.Close()
		inFile, err := os.ReadFile(inFileName)
		if err != nil {
			fmt.Printf("Failed to open input file %s: %s", inFileName, err)
			os.Exit(1)
		}
		prodos.WriteBlock(file, blockNumber, inFile)
	case "create":
		file, err := os.Create(fileName)
		if err != nil {
			fmt.Printf("failed to create file: %s\n", err)
			return
		}
		defer file.Close()
		prodos.CreateVolume(file, volumeName, volumeSize)
	case "rm":
		file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
		if err != nil {
			fmt.Printf("Failed to open drive image %s:\n  %s", fileName, err)
			os.Exit(1)
		}
		defer file.Close()
		prodos.DeleteFile(file, pathName)
	case "dumpfile":
		file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
		if err != nil {
			fmt.Printf("Failed to open drive image %s:\n  %s", fileName, err)
			os.Exit(1)
		}
		defer file.Close()
		fileEntry, err := prodos.GetFileEntry(file, pathName)
		prodos.DumpFileEntry(fileEntry)
	default:
		fmt.Printf("Invalid command: %s\n\n", command)
		flag.PrintDefaults()
		os.Exit(1)
	}
}
