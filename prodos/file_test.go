// Copyright Terence J. Boldt (c)2021-2024
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

package prodos

import (
	"fmt"
	"testing"
)

func TestCreateBlocklist(t *testing.T) {
	var tests = []struct {
		fileSize   uint32
		wantBlocks uint16
	}{
		{1, 1}, // seedling
		{512, 1},
		{513, 3}, // sapling
		{2048, 5},
		{2049, 6},
		{17128, 35},
		{131073, 260}, // tree
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
			if uint16(len(blockList)) != tt.wantBlocks {
				t.Errorf("got %d blocks, want %d", len(blockList), tt.wantBlocks)
			}
		})
	}
}
