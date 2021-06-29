package prodos

import (
	"fmt"
	"testing"
)

func TestCreateVolumeBitmap(t *testing.T) {
	var tests = []struct {
		blocks int
		want   int
	}{
		{65536, 8192},
		{65533, 8192},
		{140, 512},
	}

	for _, tt := range tests {
		testname := fmt.Sprintf("%d", tt.blocks)
		t.Run(testname, func(t *testing.T) {
			volumeBitMap := createVolumeBitmap(tt.blocks)
			ans := len(volumeBitMap)
			if ans != tt.want {
				t.Errorf("got %d, want %d", ans, tt.want)
			}
		})
	}
}

func TestCheckFreeBlockInVolumeBitmap(t *testing.T) {
	var tests = []struct {
		blocks int
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
