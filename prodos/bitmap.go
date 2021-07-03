package prodos

import (
	"os"
)

func ReadVolumeBitmap(file *os.File) []byte {
	headerBlock := ReadBlock(file, 2)

	volumeHeader := parseVolumeHeader(headerBlock)

	totalBitmapBytes := volumeHeader.TotalBlocks / 8
	if volumeHeader.TotalBlocks%8 > 0 {
		totalBitmapBytes++
	}

	bitmap := make([]byte, totalBitmapBytes)

	totalBitmapBlocks := totalBitmapBytes / 512

	if totalBitmapBytes%512 > 0 {
		totalBitmapBlocks++
	}

	for i := 0; i < totalBitmapBlocks; i++ {
		bitmapBlock := ReadBlock(file, i+volumeHeader.BitmapStartBlock)

		for j := 0; j < 512 && i*512+j < totalBitmapBytes; j++ {
			bitmap[i*512+j] = bitmapBlock[j]
		}
	}

	return bitmap
}

func writeVolumeBitmap(file *os.File, bitmap []byte) {
	headerBlock := ReadBlock(file, 2)

	volumeHeader := parseVolumeHeader(headerBlock)

	for i := 0; i < len(bitmap)/512; i++ {
		WriteBlock(file, volumeHeader.BitmapStartBlock+i, bitmap[i*512:i*512+512])
	}
}

func createVolumeBitmap(numberOfBlocks int) []byte {
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
	markBlockInVolumeBitmap(volumeBitmap, 0)
	markBlockInVolumeBitmap(volumeBitmap, 1)

	// root directory
	markBlockInVolumeBitmap(volumeBitmap, 2)
	markBlockInVolumeBitmap(volumeBitmap, 3)
	markBlockInVolumeBitmap(volumeBitmap, 4)
	markBlockInVolumeBitmap(volumeBitmap, 5)

	// volume bitmap blocks
	for i := 0; i < volumeBitmapBlocks; i++ {
		markBlockInVolumeBitmap(volumeBitmap, 6+i)
	}

	// blocks beyond the volume
	totalBlocksInBitmap := volumeBitmapBlocks * 512 * 8
	blocksBeyondEnd := totalBlocksInBitmap - numberOfBlocks
	if blocksBeyondEnd > 0 {
		for i := totalBlocksInBitmap - blocksBeyondEnd; i < totalBlocksInBitmap; i++ {
			markBlockInVolumeBitmap(volumeBitmap, i)
		}
	}
	//DumpBlock(volumeBitmap)

	return volumeBitmap
}

func findFreeBlocks(volumeBitmap []byte, numberOfBlocks int) []int {
	blocks := make([]int, numberOfBlocks)

	blocksFound := 0

	for i := 0; i < len(volumeBitmap)*8; i++ {
		if checkFreeBlockInVolumeBitmap(volumeBitmap, i) {
			blocks[blocksFound] = i
			blocksFound++
			if blocksFound == numberOfBlocks {
				return blocks
			}
		}
	}

	return nil
}

func GetFreeBlockCount(volumeBitmap []byte, totalBlocks int) int {
	freeBlockCount := 0

	for i := 0; i < totalBlocks; i++ {
		if checkFreeBlockInVolumeBitmap(volumeBitmap, i) {
			freeBlockCount++
		}
	}
	return freeBlockCount
}

func markBlockInVolumeBitmap(volumeBitmap []byte, blockNumber int) {
	bitToChange := blockNumber % 8
	byteToChange := blockNumber / 8

	byteToAnd := 0b11111111

	switch bitToChange {
	case 0:
		byteToAnd = 0b01111111
	case 1:
		byteToAnd = 0b10111111
	case 2:
		byteToAnd = 0b11011111
	case 3:
		byteToAnd = 0b11101111
	case 4:
		byteToAnd = 0b11110111
	case 5:
		byteToAnd = 0b11111011
	case 6:
		byteToAnd = 0b11111101
	case 7:
		byteToAnd = 0b11111110
	}

	//fmt.Printf("blockNumber: $%04X byteToWrite: 0b%08b volumeBitmap: $%02X byteToChange: $%04X\n", blockNumber, byteToWrite, volumeBitmap[byteToChange], byteToChange)
	volumeBitmap[byteToChange] &= byte(byteToAnd)
}

func freeBlockInVolumeBitmap(volumeBitmap []byte, blockNumber int) {
	bitToChange := blockNumber % 8
	byteToChange := blockNumber / 8

	byteToOr := 0b00000000

	switch bitToChange {
	case 0:
		byteToOr = 0b10000000
	case 1:
		byteToOr = 0b01000000
	case 2:
		byteToOr = 0b00100000
	case 3:
		byteToOr = 0b00010000
	case 4:
		byteToOr = 0b00001000
	case 5:
		byteToOr = 0b00000100
	case 6:
		byteToOr = 0b00000010
	case 7:
		byteToOr = 0b00000001
	}

	volumeBitmap[byteToChange] |= byte(byteToOr)
}

func checkFreeBlockInVolumeBitmap(volumeBitmap []byte, blockNumber int) bool {
	bitToCheck := blockNumber % 8
	byteToCheck := blockNumber / 8

	byteToAnd := 0b00000000

	switch bitToCheck {
	case 0:
		byteToAnd = 0b10000000
	case 1:
		byteToAnd = 0b01000000
	case 2:
		byteToAnd = 0b00100000
	case 3:
		byteToAnd = 0b00010000
	case 4:
		byteToAnd = 0b00001000
	case 5:
		byteToAnd = 0b00000100
	case 6:
		byteToAnd = 0b00000010
	case 7:
		byteToAnd = 0b00000001
	}

	return (volumeBitmap[byteToCheck] & byte(byteToAnd)) > 0
}
