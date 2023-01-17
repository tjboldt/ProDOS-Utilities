// Copyright Terence J. Boldt (c)2022-2023
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides access to generate a ProDOS drive image from a host directory

package prodos

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AddFilesFromHostDirectory fills the root volume with files
// from the specified host directory
func AddFilesFromHostDirectory(
	readerWriter ReaderWriterAt,
	directory string) error {

	files, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			return err
		}

		if file.Name()[0] != '.' && !file.IsDir() && info.Size() > 0 && info.Size() <= 0x1000000 {
			err = WriteFileFromFile(readerWriter, "", 0, 0, info.ModTime(), filepath.Join(directory, file.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// WriteFileFromFile writes a file to a ProDOS volume from a host file
func WriteFileFromFile(readerWriter ReaderWriterAt, pathName string, fileType int, auxType int, modifiedTime time.Time, inFileName string) error {
	fmt.Printf("WriteFileFromFile: %s\n", inFileName)
	inFile, err := os.ReadFile(inFileName)
	if err != nil {
		fmt.Println("failed to read file")
		return err
	}

	if auxType == 0 && fileType == 0 {
		auxType, fileType, inFile, err = convertFileByType(inFileName, inFile)
		if err != nil {
			fmt.Println("failed to convert file")
			return err
		}
	}

	if len(pathName) == 0 {
		_, pathName = filepath.Split(inFileName)
		pathName = strings.ToUpper(pathName)
		ext := filepath.Ext(pathName)
		if len(ext) > 0 {
			switch ext {
			case ".SYS", ".TXT", ".BAS", ".BIN":
				pathName = strings.TrimSuffix(pathName, ext)
			}
		}
	}

	return WriteFile(readerWriter, pathName, fileType, auxType, time.Now(), modifiedTime, inFile)
}

func convertFileByType(inFileName string, inFile []byte) (int, int, []byte, error) {
	fileType := 0x06  // default to BIN
	auxType := 0x2000 // default to $2000

	var err error

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
	} else {
		// use extension to determine file type
		ext := strings.ToUpper(filepath.Ext(inFileName))

		switch ext {
		case ".BAS":
			inFile, err = ConvertTextToBasic(string(inFile))
			fileType = 0xFC
			auxType = 0x0801

			if err != nil {
				return 0, 0, nil, err
			}
		case ".SYS":
			fileType = 0xFF
			auxType = 0x2000
		case ".BIN":
			fileType = 0x06
			auxType = 0x2000
		case ".TXT":
			inFile = []byte(strings.ReplaceAll(strings.ReplaceAll(string(inFile), "\r\n", "r"), "\n", "\r"))
			fileType = 0x04
			auxType = 0x0000
		case ".JPG", ".PNG":
			inFile = ConvertImageToHiResMonochrome(inFile)
			fileType = 0x06
			auxType = 0x2000
		}
	}

	return auxType, fileType, inFile, err
}
