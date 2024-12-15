package utils

import (
	"github.com/zeebo/blake3"
)

// HashChunk calculates BLAKE3 hash of a byte slice
func HashChunk(data []byte) []byte {
	hasher := blake3.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}
