// Copyright Terence J. Boldt (c)2021-2022
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides a file in memory

package prodos

type MemoryFile struct {
	data []byte
	size int
}

func NewMemoryFile(size int) *MemoryFile {
	return &MemoryFile{make([]byte, size), size}
}

func (memoryFile *MemoryFile) WriteAt(data []byte, offset int64) (int, error) {
	copy(memoryFile.data[int(offset):], data)
	return len(data), nil
}

func (memoryFile *MemoryFile) ReadAt(data []byte, offset int64) (int, error) {
	copy(data, memoryFile.data[int(offset):])
	return len(data), nil
}
