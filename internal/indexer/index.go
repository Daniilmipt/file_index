package indexer

import (
	"bytes"
	"io"
	"math"
	"os"
	"sync"

	"fileindex/pkg/types"

	"github.com/zeebo/blake3"
)

type FileIndex struct {
	sync.RWMutex
	indices map[string][]byte // filename -> hash signature
}

func NewFileIndex() *FileIndex {
	return &FileIndex{
		indices: make(map[string][]byte),
	}
}

// IndexFile indexes a binary file and stores its signature
func (fi *FileIndex) IndexFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	size := stat.Size()

	// Calculate number of chunks to sample
	numChunks := int(math.Ceil(float64(size) / float64(types.ChunkSize)))
	sampledChunks := int(float64(numChunks) * types.SampleRate)

	// Create signature
	signature := make([]byte, sampledChunks*types.HashSize)

	// Read and hash chunks
	buffer := make([]byte, types.ChunkSize)
	chunkInterval := numChunks / sampledChunks

	for i := 0; i < sampledChunks; i++ {
		offset := int64(i * chunkInterval * types.ChunkSize)
		_, err := file.Seek(offset, 0)
		if err != nil {
			return err
		}

		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		// Calculate BLAKE3 hash for chunk
		hasher := blake3.New()
		hasher.Write(buffer[:n])
		hash := hasher.Sum(nil)

		// Store hash in signature
		copy(signature[i*types.HashSize:], hash)
	}

	// Store signature in index
	fi.Lock()
	fi.indices[filepath] = signature
	fi.Unlock()

	return nil
}

// FindSimilar finds similar files to the given file
func (fi *FileIndex) FindSimilar(filepath string, errorThreshold float64) ([]string, error) {
	// Get signature of target file
	fi.RLock()
	targetSig, exists := fi.indices[filepath]
	fi.RUnlock()

	if !exists {
		return nil, os.ErrNotExist
	}

	similar := make([]string, 0)

	// Compare with all other signatures
	fi.RLock()
	for path, sig := range fi.indices {
		if path == filepath {
			continue
		}

		distance := calculateDistance(targetSig, sig)
		similarity := (types.MaxDistance - distance) / types.MaxDistance * 100

		if similarity >= (100 - errorThreshold) {
			similar = append(similar, path)
		}
	}
	fi.RUnlock()

	return similar, nil
}

// calculateDistance calculates Hamming distance between two signatures
func calculateDistance(sig1, sig2 []byte) float64 {
	if len(sig1) != len(sig2) {
		return types.MaxDistance
	}

	var distance float64
	for i := 0; i < len(sig1); i++ {
		distance += float64(bytes.Count([]byte{sig1[i] ^ sig2[i]}, []byte{1}))
	}

	return distance
}
