// Copyright Terence J. Boldt (c)2022
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides access to generate a ProDOS drive image from a host directory

package prodos

import (
	"os"
	"path/filepath"
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

		if !file.IsDir() && info.Size() > 0 && info.Size() <= 0x20000 {
			err = WriteFileFromFile(readerWriter, "", 0, 0, filepath.Join(directory, file.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
