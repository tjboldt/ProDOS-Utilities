// Copyright Terence J. Boldt (c)2021-2022
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

// This file provides a file in memory

package prodos

// MemoryFile containts file data and size
type MemoryFile struct {
	data []byte
	size int
}

// NewMemoryFile creates an in-memory file of the specified size in bytes
func NewMemoryFile(size int) *MemoryFile {
	return &MemoryFile{make([]byte, size), size}
}

// WriteAt writes data to the specified offset in the file
func (memoryFile *MemoryFile) WriteAt(data []byte, offset int64) (int, error) {
	copy(memoryFile.data[int(offset):], data)
	return len(data), nil
}

// ReadAt reads data from the specified offset in the file
func (memoryFile *MemoryFile) ReadAt(data []byte, offset int64) (int, error) {
	copy(data, memoryFile.data[int(offset):])
	return len(data), nil
}
