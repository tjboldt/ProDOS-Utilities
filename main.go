package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"

	"github.com/tjboldt/ProDOS-Utilities/prodos"
)

func main() {
	var fileName string
	var pathName string
	var command string
	var outFileName string
	flag.StringVar(&fileName, "driveimage", "", "A ProDOS format drive image")
	flag.StringVar(&pathName, "path", "", "Path name in ProDOS drive image")
	flag.StringVar(&command, "command", "ls", "Command to execute: ls, get, put, volumebitmap")
	flag.StringVar(&outFileName, "outfile", "export.bin", "Name of file to write")
	flag.Parse()

	if len(fileName) == 0 {
		fmt.Printf("Missing driveimage. Run with --help for more info.\n")
		os.Exit(1)
	}
	file, err := os.OpenFile(fileName, os.O_RDWR, 0755)
	if err != nil {
		os.Exit(1)
	}

	switch command {
	case "ls":
		volumeHeader, fileEntries := prodos.ReadDirectory(file, pathName)
		prodos.DumpDirectory(volumeHeader, fileEntries)
	case "volumebitmap":
		volumeBitmap := prodos.ReadVolumeBitmap(file)
		prodos.DumpBlock(volumeBitmap)
	case "get":
		if len(pathName) == 0 {
			fmt.Println("Missing pathname")
			os.Exit(1)
		}
		getFile := prodos.LoadFile(file, pathName)
		os.WriteFile(outFileName, getFile, fs.FileMode(os.O_RDWR))
	default:
		os.Exit(1)
	}
}
