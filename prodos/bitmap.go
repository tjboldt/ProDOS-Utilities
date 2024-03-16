// Copyright Terence J. Boldt (c)2021-2024
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides access to volum bitmap on
// a ProDOS drive image

package prodos

import (
	"io"
)

// ReadVolumeBitmap reads the volume bitmap from a ProDOS image
func ReadVolumeBitmap(reader io.ReaderAt) ([]byte, error) {
	headerBlock, err := ReadBlock(reader, 2)
	if err != nil {
		return nil, err
	}

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

	for i := uint16(0); i < totalBitmapBlocks; i++ {
		bitmapBlock, err := ReadBlock(reader, i+volumeHeader.BitmapStartBlock)
		if err != nil {
			return nil, err
		}

		for j := uint16(0); j < 512 && i*512+j < totalBitmapBytes; j++ {
			bitmap[i*512+j] = bitmapBlock[j]
		}
	}

	return bitmap, nil
}

// GetFreeBlockCount gets the number of free blocks on a ProDOS image
func GetFreeBlockCount(volumeBitmap []byte, totalBlocks uint16) uint16 {
	freeBlockCount := uint16(0)

	for i := uint16(0); i < totalBlocks; i++ {
		if checkFreeBlockInVolumeBitmap(volumeBitmap, i) {
			freeBlockCount++
		}
	}
	return freeBlockCount
}

func writeVolumeBitmap(readerWriter ReaderWriterAt, bitmap []byte) error {
	headerBlock, err := ReadBlock(readerWriter, 2)
	if err != nil {
		return err
	}

	volumeHeader := parseVolumeHeader(headerBlock)
	totalBitmapBytes := volumeHeader.TotalBlocks / 8
	if volumeHeader.TotalBlocks%8 > 0 {
		totalBitmapBytes++
	}

	totalBitmapBlocks := totalBitmapBytes / 512

	if totalBitmapBytes%512 > 0 {
		totalBitmapBlocks++
	}

	for i := uint16(0); i < totalBitmapBlocks; i++ {
		bitmapBlock, err := ReadBlock(readerWriter, i+volumeHeader.BitmapStartBlock)
		if err != nil {
			return err
		}

		for j := uint16(0); j < 512 && i*512+j < totalBitmapBytes; j++ {
			bitmapBlock[j] = bitmap[i*512+j]
		}

		err = WriteBlock(readerWriter, volumeHeader.BitmapStartBlock+i, bitmapBlock)
		if err != nil {
			return err
		}
	}

	return nil
}

func createVolumeBitmap(numberOfBlocks uint16) []byte {
	// needs to be > uint16 because it's multiplying by a uint16
	volumeBitmapBlocks := uint32(numberOfBlocks / 512 / 8)
	if volumeBitmapBlocks*8*512 < uint32(numberOfBlocks) {
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
	for i := uint32(0); i < volumeBitmapBlocks; i++ {
		markBlockInVolumeBitmap(volumeBitmap, uint16(6+i))
	}

	// blocks beyond the volume
	totalBlocksInBitmap := volumeBitmapBlocks * 512 * 8
	blocksBeyondEnd := totalBlocksInBitmap - uint32(numberOfBlocks)
	if blocksBeyondEnd > 0 {
		for i := totalBlocksInBitmap - blocksBeyondEnd; i < totalBlocksInBitmap; i++ {
			markBlockInVolumeBitmap(volumeBitmap, uint16(i))
		}
	}

	return volumeBitmap
}

func findFreeBlocks(volumeBitmap []byte, numberOfBlocks uint16) []uint16 {
	blocks := make([]uint16, numberOfBlocks)

	blocksFound := uint16(0)

	// needs to be > uint16 because it's multiplying by a uint16
	for i := uint32(0); i < uint32(len(volumeBitmap))*8; i++ {
		if checkFreeBlockInVolumeBitmap(volumeBitmap, uint16(i)) {
			blocks[blocksFound] = uint16(i)
			blocksFound++
			if blocksFound == numberOfBlocks {
				return blocks
			}
		}
	}

	return nil
}

func markBlockInVolumeBitmap(volumeBitmap []byte, blockNumber uint16) {
	bitToChange := blockNumber % 8
	byteToChange := blockNumber / 8

	byteToAnd := (uint8(0b10000000) >> uint8(bitToChange)) ^ 0b11111111

	volumeBitmap[byteToChange] &= byte(byteToAnd)
}

func freeBlockInVolumeBitmap(volumeBitmap []byte, blockNumber uint16) {
	bitToChange := blockNumber % 8
	byteToChange := blockNumber / 8

	byteToOr := uint8(0b10000000) >> uint8(bitToChange)

	volumeBitmap[byteToChange] |= byte(byteToOr)
}

func checkFreeBlockInVolumeBitmap(volumeBitmap []byte, blockNumber uint16) bool {
	bitToCheck := blockNumber % 8
	byteToCheck := blockNumber / 8

	byteToAnd := uint8(0b10000000) >> uint8(bitToCheck)

	return (volumeBitmap[byteToCheck] & byte(byteToAnd)) > 0
}
