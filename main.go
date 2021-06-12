package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tjboldt/ProDOS-Utilities/prodos"
)

func main() {
	var fileName string
	var pathName string
	var command string
	var outFileName string
	var blockNumber int
	var volumeSize int
	var volumeName string
	flag.StringVar(&fileName, "driveimage", "", "A ProDOS format drive image")
	flag.StringVar(&pathName, "path", "", "Path name in ProDOS drive image")
	flag.StringVar(&command, "command", "ls", "Command to execute: ls, get, put, volumebitmap, readblock, writeblock, createvolume, delete")
	flag.StringVar(&outFileName, "outfile", "export.bin", "Name of file to write")
	flag.IntVar(&volumeSize, "volumesize", 65535, "Number of blocks to create the volume with")
	flag.StringVar(&volumeName, "volumename", "NO.NAME", "Specifiy a name for the volume from 1 to 15 characters")
	flag.IntVar(&blockNumber, "block", 0, "A block number to read/write from 0 to 65535")
	flag.Parse()

	if len(fileName) == 0 {
		fmt.Printf("Missing driveimage. Run with --help for more info.\n")
		os.Exit(1)
	}

	switch command {
	case "ls":
		file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
		if err != nil {
			os.Exit(1)
		}
		volumeHeader, fileEntries := prodos.ReadDirectory(file, pathName)
		prodos.DumpDirectory(volumeHeader, fileEntries)
	case "volumebitmap":
		file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
		if err != nil {
			os.Exit(1)
		}
		volumeBitmap := prodos.ReadVolumeBitmap(file)
		prodos.DumpBlock(volumeBitmap)
	case "get":
		file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
		if err != nil {
			os.Exit(1)
		}
		if len(pathName) == 0 {
			fmt.Println("Missing pathname")
			os.Exit(1)
		}
		getFile := prodos.LoadFile(file, pathName)
		outFile, err := os.Create(outFileName)
		if err != nil {
			os.Exit(1)
		}
		outFile.Write(getFile)
	case "readblock":
		file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
		if err != nil {
			os.Exit(1)
		}
		block := prodos.ReadBlock(file, blockNumber)
		outFile, err := os.Create(outFileName)
		if err != nil {
			os.Exit(1)
		}
		outFile.Write(block)
	case "createvolume":
		prodos.CreateVolume(fileName, volumeName, volumeSize)
	case "delete":
		file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
		if err != nil {
			os.Exit(1)
		}
		prodos.DeleteFile(file, pathName)
	default:
		fmt.Printf("Command %s not handle\n", command)
		os.Exit(1)
	}
}
