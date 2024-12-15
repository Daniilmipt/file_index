package main

import (
	"flag"
	"fmt"
	"log"

	"fileindex/internal/indexer"
)

func main() {
	inputFile := flag.String("file", "", "File to index or search")
	searchMode := flag.Bool("search", false, "Enable search mode")
	errorThreshold := flag.Float64("error", 10.0, "Error threshold percentage (0-100)")
	flag.Parse()

	idx := indexer.NewFileIndex()

	if *searchMode {
		similar, err := idx.FindSimilar(*inputFile, *errorThreshold)
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range similar {
			fmt.Printf("Similar file found: %s\n", file)
		}
	} else {
		err := idx.IndexFile(*inputFile)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Successfully indexed: %s\n", *inputFile)
	}
}
