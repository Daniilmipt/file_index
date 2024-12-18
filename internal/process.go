package internal

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const FILE_BATCH_SIZE = 100

type HandleOptions struct {
	ErrorRate         float64
	ThreadCountIndex  int64
	ThreadCountSearch int64
}

func HandleIndex(opts HandleOptions, filePathSlice []string) {
	wg := sync.WaitGroup{}
	sem := semaphore.NewWeighted(opts.ThreadCountIndex)
	ctx := context.Background()

	for i := range filePathSlice {
		filePath := filepath.Clean(filePathSlice[i])

		wg.Add(1)
		if err := sem.Acquire(ctx, 1); err != nil {
			slog.Error("failed to acquire semaphore in index", slog.Any("error", err))
			break
		}
		go func(filePath string) {
			defer wg.Done()
			defer sem.Release(1)

			processIndex(filePath)
		}(filePath)
	}
	wg.Wait()
}

func processIndex(filePath string) {
	hash, err := Hash(filePath)
	if err != nil {
		slog.Error("Error indexing file", slog.String("error", err.Error()))
		return
	}

	fileIndex := NewFileIndex(filePath, hash)
	if err := fileIndex.SaveToFile(); err != nil {
		slog.Error("Error saving index",
			slog.String("error", err.Error()),
			slog.String("hash", fmt.Sprintf("%x", hash)),
		)
		return
	}
}

func HandleSearch(opts HandleOptions, filePathSlice []string) {
	processSearch(opts, filePathSlice)
}

func processSearch(opts HandleOptions, filePathSlice []string) {
	hashes, err := HashSlice(filePathSlice)
	if err != nil {
		slog.Error("Error indexing target file", slog.String("error", err.Error()))
		return
	}

	ch := make(chan FileIndex, FILE_BATCH_SIZE)
	eg := new(errgroup.Group)
	eg.Go(func() error {
		defer close(ch)
		return Search(
			ch,
			SearchOptions{
				Hashes:      hashes,
				ErrRate:     opts.ErrorRate,
				ThreadCount: opts.ThreadCountSearch,
			},
		)
	})

	var foundFilePaths []string
	for fi := range ch {
		foundFilePaths = append(foundFilePaths, fi.FilePath)
	}

	if err := eg.Wait(); err != nil {
		slog.Error("Error searching for similar file", slog.String("error", err.Error()))
		return
	}

	fmt.Println(foundFilePaths)
}
