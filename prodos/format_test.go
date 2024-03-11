// Copyright Terence J. Boldt (c)2021-2023
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides tests for access to format a ProDOS drive image

package prodos

import (
	"fmt"
	"testing"
)

func TestCreateVolume(t *testing.T) {
	var tests = []struct {
		blocks         uint16
		wantVolumeName string
		wantFreeBlocks uint16
	}{
		{65535, "MAX", 65513},
		{65500, "ALMOST.MAX", 65478},
		{280, "FLOPPY", 273},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.blocks)
		t.Run(testname, func(t *testing.T) {
			file := NewMemoryFile(0x2000000)

			CreateVolume(file, tt.wantVolumeName, tt.blocks)

			volumeHeader, _, fileEntries, _ := ReadDirectory(file, "")
			if volumeHeader.VolumeName != tt.wantVolumeName {
				t.Errorf("got volume name %s, want %s", volumeHeader.VolumeName, tt.wantVolumeName)
			}
			if volumeHeader.TotalBlocks != tt.blocks {
				t.Errorf("got total blocks %d, want %d", volumeHeader.TotalBlocks, tt.blocks)
			}
			if len(fileEntries) > 0 {
				t.Errorf("got files %d, want 0", len(fileEntries))
			}

			volumeBitmap, _ := ReadVolumeBitmap(file)
			freeBlockCount := GetFreeBlockCount(volumeBitmap, tt.blocks)
			if freeBlockCount != tt.wantFreeBlocks {
				t.Errorf("got free blocks: %d, want %d", freeBlockCount, tt.wantFreeBlocks)
			}
		})
	}
}
