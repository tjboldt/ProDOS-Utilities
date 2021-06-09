package prodos

import (
	"os"
)

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

	for i := 0; i < len(bitmap)/512; i++ {
		WriteBlock(file, volumeHeader.BitmapStartBlock+i, bitmap[i*512:i*512+512])
	}
}

func CreateVolumeBitmap(numberOfBlocks int) []byte {
	volumeBitmapBlocks := numberOfBlocks / 512 / 8
	if volumeBitmapBlocks*8*512 < numberOfBlocks {
		volumeBitmapBlocks++
	}

	// set all 1's to show blocks available...
	volumeBitmap := make([]byte, volumeBitmapBlocks*512)
	for i := 0; i < len(volumeBitmap); i++ {
		volumeBitmap[i] = 0xFF
	}

	// boot blocks
	MarkBlockInVolumeBitmap(volumeBitmap, 0)
	MarkBlockInVolumeBitmap(volumeBitmap, 1)

	// root directory
	MarkBlockInVolumeBitmap(volumeBitmap, 2)
	MarkBlockInVolumeBitmap(volumeBitmap, 3)
	MarkBlockInVolumeBitmap(volumeBitmap, 4)
	MarkBlockInVolumeBitmap(volumeBitmap, 5)

	// volume bitmap blocks
	for i := 0; i < volumeBitmapBlocks; i++ {
		MarkBlockInVolumeBitmap(volumeBitmap, 6+i)
	}

	// blocks beyond the volume
	totalBlocksInBitmap := volumeBitmapBlocks * 512 * 8
	blocksBeyondEnd := totalBlocksInBitmap - numberOfBlocks
	if blocksBeyondEnd > 0 {
		for i := totalBlocksInBitmap - blocksBeyondEnd; i < totalBlocksInBitmap; i++ {
			MarkBlockInVolumeBitmap(volumeBitmap, i)
		}
	}
	//DumpBlock(volumeBitmap)

	return volumeBitmap
}

func FindFreeBlocks(numberOfBlocks int) []int {
	return nil
}

func MarkBlockInVolumeBitmap(volumeBitmap []byte, blockNumber int) {
	bitToChange := blockNumber % 8
	byteToChange := blockNumber / 8

	byteToWrite := 0b11111111

	switch bitToChange {
	case 0:
		byteToWrite = 0b01111111
	case 1:
		byteToWrite = 0b10111111
	case 2:
		byteToWrite = 0b11011111
	case 3:
		byteToWrite = 0b11101111
	case 4:
		byteToWrite = 0b11110111
	case 5:
		byteToWrite = 0b11111011
	case 6:
		byteToWrite = 0b11111101
	case 7:
		byteToWrite = 0b11111110
	}

	//fmt.Printf("blockNumber: $%04X byteToWrite: 0b%08b volumeBitmap: $%02X byteToChange: $%04X\n", blockNumber, byteToWrite, volumeBitmap[byteToChange], byteToChange)
	volumeBitmap[byteToChange] &= byte(byteToWrite)
}
