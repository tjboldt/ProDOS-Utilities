// Copyright Terence J. Boldt (c)2021-2024
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides tests for access to volume bitmap on
// a ProDOS drive image

package prodos

import (
	"fmt"
	"testing"
)

func TestCreateVolumeBitmap(t *testing.T) {
	var tests = []struct {
		blocks uint16
		want   uint16
	}{
		{65535, 8192},
		{65533, 8192},
		{140, 512},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.blocks)
		t.Run(testname, func(t *testing.T) {
			volumeBitMap := createVolumeBitmap(tt.blocks)
			ans := uint16(len(volumeBitMap))
			if ans != tt.want {
				t.Errorf("got %d, want %d", ans, tt.want)
			}
		})
	}
}

func TestCheckFreeBlockInVolumeBitmap(t *testing.T) {
	var tests = []struct {
		blocks uint16
		want   bool
	}{
		{0, false},     // boot block
		{1, false},     // SOS boot block
		{2, false},     // volume root
		{21, false},    // end of volume bitmap
		{22, true},     // beginning of free space
		{8192, true},   // more free space
		{65534, true},  // last free block
		{65535, false}, // can't use last block because volume size is 0xFFFF, not 0x10000
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.blocks)
		t.Run(testname, func(t *testing.T) {
			volumeBitMap := createVolumeBitmap(65535)
			ans := checkFreeBlockInVolumeBitmap(volumeBitMap, tt.blocks)
			if ans != tt.want {
				t.Errorf("got %t, want %t", ans, tt.want)
			}
		})
	}
}

func TestMarkBlockInVolumeBitmap(t *testing.T) {
	var tests = []struct {
		blocks uint16
		want   bool
	}{
		{0, false},     // boot block
		{1, false},     // SOS boot block
		{2, false},     // volume root
		{21, false},    // end of volume bitmap
		{22, true},     // beginning of free space
		{999, false},   // end of volume bitmap
		{8192, true},   // more free space
		{65534, true},  // last free block
		{65535, false}, // can't use last block because volume size is 0xFFFF, not 0x10000
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.blocks)
		t.Run(testname, func(t *testing.T) {
			volumeBitMap := createVolumeBitmap(65535)
			markBlockInVolumeBitmap(volumeBitMap, 999)
			ans := checkFreeBlockInVolumeBitmap(volumeBitMap, tt.blocks)
			if ans != tt.want {
				t.Errorf("got %t, want %t", ans, tt.want)
			}
		})
	}
}

func TestUpdateVolumeBitmap(t *testing.T) {
	blockList := []uint16{10, 11, 12, 100, 120}

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
