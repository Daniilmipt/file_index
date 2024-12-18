package internal

import (
	"fmt"
	"io"
	"os"

	"github.com/zeebo/blake3"
)

const (
	INDEX_FILES_DIR = "indexed_files"
	BATCH_SIZE      = 1024
)

// IsSimilarHash compares two hashes with a similarity threshold
func IsSimilarHash(hash1, hash2 []byte, errorRate float64) bool {
	distance := hammingDistance(hash1, hash2)
	return float64(distance)/float64(len(hash1)*8) <= errorRate
}

// hammingDistance computes the Hamming distance between two hashes
func hammingDistance(hash1, hash2 []byte) int {
	var distance int
	for i := 0; i < len(hash1); i++ {
		distance += popCount(hash1[i] ^ hash2[i])
	}
	return distance
}

// popCount counts bits set to 1
func popCount(x byte) int {
	count := 0
	for x > 0 {
		count++
		x &= x - 1
	}
	return count
}

// Hash computes the hash of a file
func Hash(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	hasher := blake3.New()
	if _, err = io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("failed to compress file: %w", err)
	}

	return hasher.Sum(nil), nil
}

func HashSlice(filePaths []string) ([][]byte, error) {
	hashes := make([][]byte, 0, len(filePaths))

	for i := range filePaths {
		hash, err := Hash(filePaths[i])
		if err != nil {
			return nil, fmt.Errorf("failed to compress file: %w", err)
		}

		hashes = append(hashes, hash)
	}
	return hashes, nil
}
