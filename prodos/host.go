// Copyright Terence J. Boldt (c)2022-2024
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides access to generate a ProDOS drive image from a host directory

package prodos

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// AddFilesFromHostDirectory fills the root volume with files
// from the specified host directory
func AddFilesFromHostDirectory(
	readerWriter ReaderWriterAt,
	directory string,
	path string,
	recursive bool,
) error {

	path, err := makeFullPath(path, readerWriter)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}

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
			err = WriteFileFromFile(readerWriter, path, 0, 0, info.ModTime(), filepath.Join(directory, file.Name()), true)
			if err != nil {
				return err
			}
		}

		if file.Name()[0] != '.' && recursive && file.IsDir() {
			newPath := file.Name()
			if len(newPath) > 15 {
				newPath = newPath[0:15]
			}
			newFullPath := strings.ToUpper(path + newPath)

			newHostDirectory := filepath.Join(directory, file.Name())
			err = CreateDirectory(readerWriter, newFullPath)
			if err != nil {
				return err
			}
			err = AddFilesFromHostDirectory(readerWriter, newHostDirectory, newFullPath+"/", recursive)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// WriteFileFromFile writes a file to a ProDOS volume from a host file
func WriteFileFromFile(
	readerWriter ReaderWriterAt,
	pathName string,
	fileType uint8,
	auxType uint16,
	modifiedTime time.Time,
	inFileName string,
	ignoreDuplicates bool,
) error {

	inFile, err := os.ReadFile(inFileName)
	if err != nil {
		errString := fmt.Sprintf("write from file failed: %s", err)
		return errors.New(errString)
	}

	if auxType == 0 && fileType == 0 {
		auxType, fileType, inFile, err = convertFileByType(inFileName, inFile)
		if err != nil {
			errString := fmt.Sprintf("failed to convert file: %s", err)
			return errors.New(errString)
		}
	}

	trimExtensions := false
	if len(pathName) == 0 {
		_, pathName = filepath.Split(inFileName)
		pathName = strings.ToUpper(pathName)
		trimExtensions = true
	}

	if strings.HasSuffix(pathName, "/") {
		trimExtensions = true
		_, fileName := filepath.Split(inFileName)
		pathName = strings.ToUpper(pathName + fileName)
	}

	if trimExtensions {
		ext := filepath.Ext(pathName)

		if len(ext) > 0 {
			switch ext {
			case ".SYS", ".TXT", ".BAS", ".BIN":
				pathName = strings.TrimSuffix(pathName, ext)
			}
			match, err := regexp.MatchString("^\\.(BIN|SYS|TXT|BAS)\\$[0-9]{4}", ext)

			if err == nil && match {
				pathName = strings.TrimSuffix(pathName, ext)
			}
		}
	}

	paths := strings.SplitAfter(pathName, "/")
	if len(paths[len(paths)-1]) > 15 {
		paths[len(paths)-1] = paths[len(paths)-1][0:15]
		pathName = strings.Join(paths, "")
	}

	// skip if file already exists and ignoring duplicates
	if ignoreDuplicates {
		exists, err := FileExists(readerWriter, pathName)
		if err != nil {
			return err
		}
		if exists {
			return nil
		}
	}

	return WriteFile(readerWriter, pathName, fileType, auxType, time.Now(), modifiedTime, inFile)
}

func convertFileByType(inFileName string, inFile []byte) (uint16, uint8, []byte, error) {
	var auxType uint16
	var fileType uint8

	fileType = 0x06  // default to BIN
	auxType = 0x2000 // default to $2000

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

		fileType = uint8(binary.BigEndian.Uint16(inFile[0x34:]))
		auxType = uint16(binary.BigEndian.Uint32(inFile[0x36:]))
		inFile = inFile[0x3A:]
	} else {
		// use extension to determine file type
		ext := strings.ToUpper(filepath.Ext(inFileName))

		match, err := regexp.MatchString("^\\.(BIN|SYS|TXT|BAS)\\$[0-9]{4}", ext)

		if err == nil && match {
			parts := strings.Split(ext, "$")
			extAuxType, err := strconv.ParseUint(parts[1], 16, 16)
			if err != nil {
				return 0, 0, nil, err
			}
			auxType = uint16(extAuxType)
			switch parts[0] {
			case ".BAS":
				fileType = 0xFC
			case ".SYS":
				fileType = 0xFF
			case ".BIN":
				fileType = 0x06
			case ".TXT":
				fileType = 0x04
			}
		} else {
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
	}

	return auxType, fileType, inFile, err
}
