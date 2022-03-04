package prodos

import (
	"fmt"
	"testing"
)

func TestCreateVolume(t *testing.T) {
	var tests = []struct {
		blocks         int
		wantVolumeName string
		wantFreeBlocks int
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
