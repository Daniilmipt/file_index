package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"fileindex/cmd/config"
	"fileindex/internal"
)

const (
	CONFIG_PATH_ENV = "CONFIG_PATH"
)

func main() {
	cfg := config.InitConfig(CONFIG_PATH_ENV)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	command := flag.String("command", "", "can be index if you want to index a file, or search if you want to search for a similar file")
	filePath := flag.String("filepath", "", "list of files you want to index or search for. pass in formt \"<filename1>,<filename2>\"")

	flag.Parse()

	if command == nil || filePath == nil {
		slog.Error("Empty command or filepath in command line arguments",
			slog.Any("command", command),
			slog.Any("filepath", filePath),
		)
	}

	filePathSlice := strings.Split(*filePath, ",")

	switch *command {
	case "index":
		internal.HandleIndex(
			internal.HandleOptions{
				ThreadCountIndex: cfg.General.ThreadCountIndex,
			},
			filePathSlice,
		)
	case "search":
		internal.HandleSearch(
			internal.HandleOptions{
				ErrorRate:         cfg.General.ErrorRate,
				ThreadCountSearch: cfg.General.ThreadCountSearch,
			},
			filePathSlice,
		)
	default:
		slog.Error("Unknown command", slog.Any("command", *command))
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage: <command> [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  index=<file_path>    Index a file")
	fmt.Println("  search=<file_path>   Search for a similar file")
}
