// Copyright Terence J. Boldt (c)2021-2023
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides a command line utility to read, write and delete
// files and directories on a ProDOS drive image as well as format
// new volumes

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tjboldt/ProDOS-Utilities/prodos"
)

const version = "0.4.3"

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
	flag.StringVar(&command, "c", "ls", "Command to execute: ls, get, put, rm, mkdir, readblock, writeblock, create, putall")
	flag.StringVar(&outFileName, "o", "", "Name of file to write")
	flag.StringVar(&inFileName, "i", "", "Name of file to read")
	flag.IntVar(&volumeSize, "s", 65535, "Number of blocks to create the volume with (default 65535, 64 to 65535, 0x0040 to 0xFFFF hex input accepted)")
	flag.StringVar(&volumeName, "v", "NO.NAME", "Specifiy a name for the volume from 1 to 15 characters")
	flag.IntVar(&blockNumber, "b", 0, "A block number to read/write from 0 to 65535 (0x0000 to 0xFFFF hex input accepted)")
	flag.IntVar(&fileType, "t", 0, "ProDOS FileType: 0x04 for TXT, 0x06 for BIN, 0xFC for BAS, 0xFF for SYS etc., omit to autodetect")
	flag.IntVar(&auxType, "a", 0, "ProDOS AuxType from 0 to 65535 (0x0000 to 0xFFFF hex input accepted), omit to autodetect")
	flag.Parse()

	if len(fileName) == 0 {
		printReadme()
		flag.PrintDefaults()
		os.Exit(1)
	}

	switch command {
	case "ls":
		ls(fileName, pathName)
	case "get":
		get(fileName, pathName, outFileName)
	case "put":
		put(fileName, pathName, fileType, auxType, inFileName)
	case "readblock":
		readBlock(blockNumber, fileName)
	case "writeblock":
		writeBlock(blockNumber, fileName, inFileName)
	case "create":
		create(fileName, volumeName, volumeSize)
	case "putall":
		putall(fileName, inFileName)
	case "rm":
		rm(fileName, pathName)
	case "mkdir":
		mkdir(fileName, pathName)
	case "dumpfile":
		dumpFile(fileName, pathName)
	default:
		fmt.Printf("Invalid command: %s\n\n", command)
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func dumpFile(fileName string, pathName string) {
	checkPathName(pathName)
	file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
	if err != nil {
		fmt.Printf("Failed to open drive image %s:\n  %s", fileName, err)
		os.Exit(1)
	}
	defer file.Close()
	fileEntry, err := prodos.GetFileEntry(file, pathName)
	prodos.DumpFileEntry(fileEntry)
}

func mkdir(fileName string, pathName string) {
	checkPathName(pathName)
	file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
	if err != nil {
		fmt.Printf("Failed to open drive image %s:\n  %s", fileName, err)
		os.Exit(1)
	}
	defer file.Close()
	err = prodos.CreateDirectory(file, pathName)
	if err != nil {
		fmt.Printf("failed to create directory %s: %s\n", pathName, err)
		os.Exit(1)
	}
}

func rm(fileName string, pathName string) {
	checkPathName(pathName)
	file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
	if err != nil {
		fmt.Printf("Failed to open drive image %s:\n  %s", fileName, err)
		os.Exit(1)
	}
	defer file.Close()
	prodos.DeleteFile(file, pathName)
}

func putall(fileName string, inFileName string) {
	if len(inFileName) == 0 {
		inFileName = "."
	}
	file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
	if err != nil {
		fmt.Printf("failed to create file: %s\n", err)
		os.Exit(1)
	}
	defer file.Close()
	err = prodos.AddFilesFromHostDirectory(file, inFileName)
	if err != nil {
		fmt.Printf("failed to add host files: %s\n", err)
		os.Exit(1)
	}
}

func create(fileName string, volumeName string, volumeSize int) {
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("failed to create file: %s\n", err)
		os.Exit(1)
	}
	defer file.Close()
	prodos.CreateVolume(file, volumeName, volumeSize)
}

func writeBlock(blockNumber int, fileName string, inFileName string) {
	checkInFileName(inFileName)
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
}

func readBlock(blockNumber int, fileName string) {
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
}

func put(fileName string, pathName string, fileType int, auxType int, inFileName string) {
	checkPathName(pathName)
	checkInFileName(inFileName)
	file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
	if err != nil {
		fmt.Printf("Failed to open drive image %s:\n  %s", fileName, err)
		os.Exit(1)
	}
	defer file.Close()
	fileInfo, err := os.Stat(fileName)

	err = prodos.WriteFileFromFile(file, pathName, fileType, auxType, fileInfo.ModTime(), inFileName)
	if err != nil {
		fmt.Printf("Failed to write file %s", err)
	}
}

func get(fileName string, pathName string, outFileName string) {
	checkPathName(pathName)
	file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
	if err != nil {
		fmt.Printf("Failed to open drive image %s:\n  %s", fileName, err)
		os.Exit(1)
	}
	defer file.Close()
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
}

func ls(fileName string, pathName string) {
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
}

func checkPathName(pathName string) {
	if len(pathName) == 0 {
		fmt.Printf("Missing path name (use -p PATHNAME)\n")
		os.Exit(1)
	}
}

func checkInFileName(inFileName string) {
	if len(inFileName) == 0 {
		fmt.Printf("Missing input file name (use -i FILENAME)\n")
		os.Exit(1)
	}
}
