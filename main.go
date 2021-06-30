package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tjboldt/ProDOS-Utilities/prodos"
)

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
	flag.StringVar(&fileName, "driveimage", "", "A ProDOS format drive image")
	flag.StringVar(&pathName, "path", "", "Path name in ProDOS drive image")
	flag.StringVar(&command, "command", "ls", "Command to execute: ls, get, put, volumebitmap, readblock, writeblock, createvolume, delete")
	flag.StringVar(&outFileName, "outfile", "", "Name of file to write")
	flag.StringVar(&inFileName, "infile", "", "Name of file to read")
	flag.IntVar(&volumeSize, "volumesize", 65535, "Number of blocks to create the volume with")
	flag.StringVar(&volumeName, "volumename", "NO.NAME", "Specifiy a name for the volume from 1 to 15 characters")
	flag.IntVar(&blockNumber, "block", 0, "A block number to read/write from 0 to 65535")
	flag.IntVar(&fileType, "type", 6, "ProDOS FileType: 4=txt, 6=bin, 252=bas, 255=sys etc.")
	flag.IntVar(&auxType, "aux", 0x2000, "ProDOS AuxType from 0 to 65535 (usually load address)")
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
		volumeHeader, _, fileEntries := prodos.ReadDirectory(file, pathName)
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
		getFile, err := prodos.LoadFile(file, pathName)
		if err != nil {
			fmt.Printf("Failed to read file %s: %s", pathName, err)
		}
		if len(outFileName) == 0 {
			_, outFileName = prodos.GetDirectoryAndFileNameFromPath(pathName)
		}
		outFile, err := os.Create(outFileName)
		if err != nil {
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
			os.Exit(1)
		}
		if len(pathName) == 0 {
			fmt.Println("Missing pathname")
			os.Exit(1)
		}
		inFile, err := os.ReadFile(inFileName)
		if err != nil {
			os.Exit(1)
		}
		err = prodos.WriteFile(file, pathName, fileType, auxType, inFile)
		if err != nil {
			fmt.Printf("Failed to write file %s: %s", pathName, err)
		}
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
