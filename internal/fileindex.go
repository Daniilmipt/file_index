package internal

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"golang.org/x/sync/semaphore"
)

const (
	INDEX_FILE_NAME = "file_index"
)

type FileIndex struct {
	PathLength uint32
	HashLength uint32
	FilePath   string
	Hash       []byte
}

// NewFileIndex creates a new FileIndex
func NewFileIndex(filePath string, hash []byte) *FileIndex {
	return &FileIndex{
		PathLength: uint32(len(filePath)),
		HashLength: uint32(len(hash)),
		FilePath:   filePath,
		Hash:       hash,
	}
}

// SaveToFile saves the index by appending to the binary file
func (fi *FileIndex) SaveToFile() error {
	file, err := os.OpenFile(INDEX_FILE_NAME, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write PathLength and HashLength
	if err := binary.Write(file, binary.LittleEndian, fi.PathLength); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, fi.HashLength); err != nil {
		return err
	}

	// Write FilePath as bytes
	if _, err := file.Write([]byte(fi.FilePath)); err != nil {
		return err
	}

	// Write Hash
	if _, err := file.Write(fi.Hash); err != nil {
		return err
	}

	return nil
}

type SearchOptions struct {
	Hashes      [][]byte
	ErrRate     float64
	ThreadCount int64
}

type syncOptions struct {
	wg  *sync.WaitGroup
	sem *semaphore.Weighted
}

// Search searches for a similar file in the index
func Search(ch chan<- FileIndex, searchOpts SearchOptions) (res error) {
	file, err := os.Open(INDEX_FILE_NAME)
	if err != nil {
		return fmt.Errorf("cannot open index file: %w", err)
	}
	defer file.Close()

	ctx := context.Background()
	syncOpts := syncOptions{
		wg:  &sync.WaitGroup{},
		sem: semaphore.NewWeighted(searchOpts.ThreadCount),
	}

	for {
		entry, err := readBinaryFile(file)
		if err != nil {
			if err != io.EOF {
				res = fmt.Errorf("failed read index file: %w", err)
			}
			break
		}
		searchSimilarHash(ctx, &entry, searchOpts, syncOpts, ch)
	}
	syncOpts.wg.Wait()
	return res
}

func searchSimilarHash(
	ctx context.Context,
	entry *FileIndex,
	searchOpts SearchOptions, syncOpts syncOptions,
	ch chan<- FileIndex,
) {
	for i := range searchOpts.Hashes {
		hash := searchOpts.Hashes[i]

		syncOpts.wg.Add(1)
		if err := syncOpts.sem.Acquire(ctx, 1); err != nil {
			slog.Error("failed to acquire semaphore in search", slog.Any("error", err))
			return
		}
		go func() {
			defer syncOpts.wg.Done()
			defer syncOpts.sem.Release(1)

			if IsSimilarHash(hash, entry.Hash, searchOpts.ErrRate) {
				if _, err := os.Stat(entry.FilePath); err != nil {
					return
				}
				ch <- *entry
			}
		}()
	}
}

func readBinaryFile(file *os.File) (FileIndex, error) {
	var fi FileIndex

	// Read PathLength and HashLength
	if err := binary.Read(file, binary.LittleEndian, &fi.PathLength); err != nil {
		return fi, err
	}
	if err := binary.Read(file, binary.LittleEndian, &fi.HashLength); err != nil {
		return fi, err
	}

	// Read FilePath
	pathBuf := make([]byte, fi.PathLength)
	if _, err := io.ReadFull(file, pathBuf); err != nil {
		return fi, err
	}
	fi.FilePath = string(pathBuf)

	// Read Hash
	hashBuf := make([]byte, fi.HashLength)
	if _, err := io.ReadFull(file, hashBuf); err != nil {
		return fi, err
	}
	fi.Hash = hashBuf

	return fi, nil
}
