package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/klauspost/compress/gzip"
	"github.com/zeebo/blake3"
)

const (
	indexedFilesDir = "indexed_files" // Directory for storing files by hash
	batchSize       = 1024 // Batch size for reading and writing files
)

// IndexFile computes the hash of a file
func IndexFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	hasher := blake3.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("cannot compute hash: %w", err)
	}
	return hasher.Sum(nil), nil
}

// CopyFileToHashFile copies a file to a new file with a hash as the name
func CopyFileToHashFile(filePath string, hash []byte) (string, error) {
	hashStr := fmt.Sprintf("%x", hash)
	newFilePath := filepath.Join(indexedFilesDir, hashStr)

	if err := os.MkdirAll(indexedFilesDir, 0755); err != nil {
		return "", fmt.Errorf("cannot create directory: %w", err)
	}

	originalFile, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot open original file: %w", err)
	}
	defer originalFile.Close()

	newFile, err := os.Create(newFilePath)
	if err != nil {
		return "", fmt.Errorf("cannot create new file: %w", err)
	}
	defer newFile.Close()

	gw := gzip.NewWriter(newFile)
	defer gw.Close()

	if err := writeFileToHashFile(originalFile, gw); err != nil {
		return "", fmt.Errorf("error writing file to hash file: %w", err)
	}

	if err := gw.Close(); err != nil {
		return "", fmt.Errorf("error closing gzip writer: %w", err)
	}

	return newFilePath, nil
}

func writeFileToHashFile(file *os.File, gw *gzip.Writer) error {
	buffer := make([]byte, batchSize)
	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading file: %w", err)
		}

		if _, err = gw.Write(buffer[:n]); err != nil {
			return fmt.Errorf("error writing compressed data: %w", err)
		}
	}
	return nil
}
