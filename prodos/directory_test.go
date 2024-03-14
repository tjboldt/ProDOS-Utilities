// Copyright Terence J. Boldt (c)2024
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides tests for directories on
// a ProDOS drive image

package prodos

import (
	"testing"
)

func TestCreateDirectoryWithoutPathFails(t *testing.T) {
	t.Run("TestCreateDirectoryWithoutPathFails", func(t *testing.T) {
		file := NewMemoryFile(0x2000000)

		CreateVolume(file, "test.volume", 1024)

		err := CreateDirectory(file, "")

		if err == nil {
			t.Errorf("got nil, want non-nil")
		}
	})
}

func TestCreateDuplicateDirectoryFails(t *testing.T) {
	t.Run("TestCreateDuplicateDirectoryFails", func(t *testing.T) {
		file := NewMemoryFile(0x2000000)

		CreateVolume(file, "test.volume", 1024)

		err := CreateDirectory(file, "duplicate")
		if err != nil {
			t.Errorf("failed to create directory: %s", err)
		}

		err = CreateDirectory(file, "duplicate")

		if err == nil {
			t.Error("got nil, want non-nil")
		}
	})
}

func TestCreateAndReadDirectory(t *testing.T) {
	var tests = []struct {
		testName      string
		createPath    string
		readPath      string
		expectedCount int
	}{
		{"checkRoot", "", "/test", 0},
		{"checkCreateInRoot", "one", "/test", 1},
		{"checkRootCreateInSub", "/test/one/two", "/test", 1},
		{"checkSub", "", "/test/one", 1},
	}

	file := NewMemoryFile(0x2000000)
	CreateVolume(file, "test", 1024)

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			if len(tt.createPath) > 0 {
				CreateDirectory(file, tt.createPath)
			}
			_, _, fileEntries, err := ReadDirectory(file, tt.readPath)

			if err != nil {
				t.Errorf("got error %s", err)
			}
			got := len(fileEntries)
			if got != int(tt.expectedCount) {
				t.Errorf("got %d, want %d", got, tt.expectedCount)
			}
		})
	}
}
