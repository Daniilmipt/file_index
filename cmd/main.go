package main

import (
	"fmt"
	"os"

	"fileindex/internal/core"
	"fileindex/internal/storage"
)

const (
	HashErrorMargin = 0.1 // Error rate for hash comparison (10%)
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	fileIndex := core.NewFileIndex()

	switch command {
	case "index":
		handleIndex(fileIndex)
	case "search":
		handleSearch(fileIndex)
	default:
		fmt.Println("Unknown command:", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: <command> [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  index <file_path>    Index a file")
	fmt.Println("  search <file_path>   Search for a similar file")
}

func handleIndex(fileIndex *core.FileIndex) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: index <file_path>")
		return
	}
	filePath := os.Args[2]

	hash, err := storage.IndexFile(filePath)
	if err != nil {
		fmt.Println("Error indexing file:", err)
		return
	}

	newFilePath, err := storage.CopyFileToHashFile(filePath, hash)
	if err != nil {
		fmt.Println("Error copying file:", err)
		return
	}

	fileIndex.Set(newFilePath, hash)
	if err := fileIndex.SaveToFile(); err != nil {
		fmt.Println("Error saving index:", err)
		return
	}

	fmt.Printf("File indexed: %s\n", filePath)
}

func handleSearch(fileIndex *core.FileIndex) {
	if len(os.Args) < 3 {
		fmt.Println("Usage: search <file_path>")
		return
	}
	filePath := os.Args[2]

	targetHash, err := storage.IndexFile(filePath)
	if err != nil {
		fmt.Println("Error indexing target file:", err)
		return
	}

	entry, err := fileIndex.Search(targetHash, HashErrorMargin)
	if err!= nil {
		fmt.Println("Error searching index:", err)
		return
	}
	fmt.Printf("Found similar file: %s\n", entry.FilePath)
}
