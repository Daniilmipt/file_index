package core

import (
	"encoding/binary"
	"fileindex/internal/config"
	"fmt"
	"io"
	"os"
)

type FileIndex struct {
	PathLength uint32 // Length of the file path
	FilePath   string // Actual file path
	HashLength uint32 // Length of the hash
	Hash       []byte // Actual hash
}

// NewFileIndex creates a new FileIndex
func NewFileIndex() *FileIndex {
	return &FileIndex{}
}

// Set sets the file path and hash for this entry
func (fi *FileIndex) Set(filePath string, hash []byte) {
	fi.FilePath = filePath
	fi.Hash = hash
	fi.PathLength = uint32(len(filePath))
	fi.HashLength = uint32(len(hash))
}

// SaveToFile saves the index by appending to the binary file
func (fi *FileIndex) SaveToFile() error {
	file, err := os.OpenFile(config.IndexFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write path length and path
	fi.PathLength = uint32(len(fi.FilePath))
	if err := binary.Write(file, binary.LittleEndian, fi.PathLength); err != nil {
		return err
	}
	if _, err := file.Write([]byte(fi.FilePath)); err != nil {
		return err
	}

	// Write hash length and hash
	fi.HashLength = uint32(len(fi.Hash))
	if err := binary.Write(file, binary.LittleEndian, fi.HashLength); err != nil {
		return err
	}
	if _, err := file.Write(fi.Hash); err != nil {
		return err
	}

	return nil
}

// Search searches for a similar file in the index
func (fi *FileIndex) Search(targetHash []byte, errorRate float64) (*FileIndex, error) {
	file, err := os.Open(config.IndexFileName)
	if err != nil {
		return nil, fmt.Errorf("cannot open index file: %w", err)
	}
	defer file.Close()

	for {
		var entry FileIndex

		// Read path length
		if err := binary.Read(file, binary.LittleEndian, &entry.PathLength); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("error reading path length: %w", err)
		}

		// Read path
		pathBytes := make([]byte, entry.PathLength)
		if _, err := io.ReadFull(file, pathBytes); err != nil {
			return nil, fmt.Errorf("error reading path: %w", err)
		}
		entry.FilePath = string(pathBytes)

		// Read hash length
		if err := binary.Read(file, binary.LittleEndian, &entry.HashLength); err != nil {
			return nil, fmt.Errorf("error reading hash length: %w", err)
		}

		// Read hash
		entry.Hash = make([]byte, entry.HashLength)
		if _, err := io.ReadFull(file, entry.Hash); err != nil {
			return nil, fmt.Errorf("error reading hash: %w", err)
		}

		if IsSimilar(entry.Hash, targetHash, errorRate) {
			// Verify the file still exists at the original path
			if _, err := os.Stat(entry.FilePath); err == nil {
				return &entry, nil
			}
			continue // Skip if file doesn't exist anymore
		}
	}

	return nil, fmt.Errorf("no similar file found")
}
