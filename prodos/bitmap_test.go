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
			volumeBitMap := CreateVolumeBitmap(tt.blocks)
			ans := len(volumeBitMap)
			if ans != tt.want {
				t.Errorf("got %d, want %d", ans, tt.want)
			}
		})
	}
}
