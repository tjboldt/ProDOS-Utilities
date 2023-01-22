// Copyright Terence J. Boldt (c)2021-2023
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package prodos

import (
	"fmt"
	"testing"
)

func TestCreateBlocklist(t *testing.T) {
	var tests = []struct {
		fileSize   int
		wantBlocks int
	}{
		{1, 1},
		{512, 1},
		{513, 3},
		{2048, 5},
		{2049, 6},
		{17128, 35},
	}

	virtualDisk := NewMemoryFile(0x2000000)
	CreateVolume(virtualDisk, "VIRTUAL.DISK", 0xFFFE)

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.fileSize)
		t.Run(testname, func(t *testing.T) {
			blockList, err := createBlockList(virtualDisk, tt.fileSize)

			if err != nil {
				t.Error("got error, want nil")
			}
			if len(blockList) != tt.wantBlocks {
				t.Errorf("got %d blocks, want %d", len(blockList), tt.wantBlocks)
			}
		})
	}
}

func TestUpdateVolumeBitmap(t *testing.T) {
	blockList := []int{10, 11, 12, 100, 120}

	virtualDisk := NewMemoryFile(0x2000000)
	CreateVolume(virtualDisk, "VIRTUAL.DISK", 0xFFFE)
	updateVolumeBitmap(virtualDisk, blockList)

	for _, tt := range blockList {
		testname := fmt.Sprintf("%d", tt)
		t.Run(testname, func(t *testing.T) {

			volumeBitmap, err := ReadVolumeBitmap(virtualDisk)
			if err != nil {
				t.Error("got error, want nil")
			}
			free := checkFreeBlockInVolumeBitmap(volumeBitmap, tt)
			if free {
				t.Errorf("got true, want false")
			}
		})
	}
}
