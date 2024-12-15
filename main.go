package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/zeebo/blake3"
)

// isSimilar compares two hashes with a similarity threshold
func isSimilar(hash1, hash2 []byte, errorRate float64) bool {
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

const (
	indexFileName   = "file_index.gob" // Binary format for the index
	storageDir      = "storage"        // Directory for storing files by hash
	hashErrorMargin = 0.1              // Error rate for hash comparison (10%)
)

type FileEntry struct {
	FilePath string `json:"file_path"`
	Hash     []byte `json:"hash"`
}

type FileIndex struct {
	Entries []FileEntry `json:"entries"`
}

// NewFileIndex creates a new FileIndex
func NewFileIndex() *FileIndex {
	return &FileIndex{Entries: []FileEntry{}}
}

// Add adds a new file entry to the index
func (fi *FileIndex) Add(filePath string, hash []byte) {
	fi.Entries = append(fi.Entries, FileEntry{
		FilePath: filePath,
		Hash:     hash,
	})
}

// SaveToFile saves the index to a binary file
func (fi *FileIndex) SaveToFile() error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(fi); err != nil {
		return err
	}
	return os.WriteFile(indexFileName, buf.Bytes(), 0644)
}

// LoadFromFile loads the index from a binary file
func (fi *FileIndex) LoadFromFile() error {
	data, err := os.ReadFile(indexFileName)
	if err != nil {
		return err
	}
	decoder := gob.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(fi)
}

func (fi *FileIndex) Search(targetHash []byte, errorRate float64) *FileEntry {
	for _, entry := range fi.Entries {
		if isSimilar(entry.Hash, targetHash, errorRate) {
			// Verify the file still exists at the original path
			if _, err := os.Stat(entry.FilePath); err == nil {
				return &entry
			}
			fmt.Printf("File not found: %s\n", entry.FilePath)
		}
	}
	return nil
}

func IndexFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	// Compute the hash
	hasher := blake3.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return nil, fmt.Errorf("cannot compute hash: %w", err)
	}
	hash := hasher.Sum(nil)

	return hash, nil
}

func copyFileToHashFile(filePath string, hash []byte) (string, error) {
	// Convert hash to a string (e.g., hexadecimal)
	hashStr := fmt.Sprintf("%x", hash)

	// Define the path for the new file (name is the hash)
	newFilePath := filepath.Join("indexed_files", hashStr)

	// Ensure the target directory exists
	if err := os.MkdirAll("indexed_files", 0755); err != nil {
		return "", fmt.Errorf("cannot create directory: %w", err)
	}

	// Open the original file
	originalFile, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot open original file: %w", err)
	}
	defer originalFile.Close()

	// Create the new file with the hash name
	newFile, err := os.Create(newFilePath)
	if err != nil {
		return "", fmt.Errorf("cannot create new file: %w", err)
	}
	defer newFile.Close()

	// Copy the content from the original file to the new file
	if _, err := io.Copy(newFile, originalFile); err != nil {
		return "", fmt.Errorf("cannot copy file content: %w", err)
	}

	return newFilePath, nil
}

func readFileContent(filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}
	return content, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: <command> [arguments]")
		fmt.Println("Commands:")
		fmt.Println("  index <file_path>    Index a file")
		fmt.Println("  search <file_path>   Search for a similar file")
		os.Exit(1)
	}

	command := os.Args[1]
	fileIndex := NewFileIndex()

	// Load the index if it exists
	if _, err := os.Stat(indexFileName); err == nil {
		if err := fileIndex.LoadFromFile(); err != nil {
			fmt.Println("Error loading index:", err)
			os.Exit(1)
		}
	}

	switch command {
	case "index":
		if len(os.Args) < 3 {
			fmt.Println("Usage: index <file_path>")
			os.Exit(1)
		}
		filePath := os.Args[2]

		hash, err := IndexFile(filePath)
		if err != nil {
			fmt.Println("Error indexing file:", err)
			os.Exit(1)
		}

		// Copy the file content to a new file named after the hash
		newFilePath, err := copyFileToHashFile(filePath, hash)
		if err != nil {
			fmt.Println("Error copying file:", err)
			os.Exit(1)
		}

		fileIndex.Add(newFilePath, hash)
		if err := fileIndex.SaveToFile(); err != nil {
			fmt.Println("Error saving index:", err)
			os.Exit(1)
		}
		fmt.Printf("File indexed: %s\n", filePath)

	case "search":
		if len(os.Args) < 3 {
			fmt.Println("Usage: search <file_path>")
			os.Exit(1)
		}
		filePath := os.Args[2]

		targetHash, err := IndexFile(filePath)
		if err != nil {
			fmt.Println("Error indexing target file:", err)
			os.Exit(1)
		}

		entry := fileIndex.Search(targetHash, hashErrorMargin)
		if entry != nil {
			fmt.Printf("Found similar file: %s\n", entry.FilePath)
		} else {
			fmt.Println("No similar file found.")
		}

	default:
		fmt.Println("Unknown command:", command)
		fmt.Println("Commands: index, search")
		os.Exit(1)
	}
}
