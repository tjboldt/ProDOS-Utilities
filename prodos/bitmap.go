package prodos

import "os"

func ReadVolumeBitmap(file *os.File) []byte {
	headerBlock := ReadBlock(file, 2)

	volumeHeader := parseVolumeHeader(headerBlock)

	bitmap := make([]byte, volumeHeader.TotalBlocks/8+1)

	totalBitmapBlocks := volumeHeader.TotalBlocks / 8 / 512

	for i := 0; i <= totalBitmapBlocks; i++ {
		bitmapBlock := ReadBlock(file, i+volumeHeader.BitmapStartBlock)

		for j := 0; j < 512; j++ {
			bitmap[i*512+j] = bitmapBlock[j]
		}
	}

	return bitmap
}

func WriteVolumeBitmap(file *os.File, bitmap []byte) {
	headerBlock := ReadBlock(file, 2)

	volumeHeader := parseVolumeHeader(headerBlock)

	for i := 0; i < len(bitmap)/512/8; i++ {
		WriteBlock(file, volumeHeader.BitmapStartBlock+1, bitmap[i*512:i*512+513])
	}
}

func FindFreeBlocks(numberOfBlocks int) []int {
	return nil
}
