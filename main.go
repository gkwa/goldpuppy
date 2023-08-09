package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileLink struct {
	FilePath string   `json:"file_path"`
	Symlinks []string `json:"symlinks"`
}

var (
	debug    bool
	skipDirs string
)

func main() {
	var filePaths []string
	var outputFile string

	// Incorporate Golang flags
	flag.StringVar(&outputFile, "output", "goldpuppy.json", "Path to the output JSON file")
	flag.StringVar(&skipDirs, "skipDirs", "/proc", "Comma-separated list of directories to skip while walking the file system")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
	flag.Parse()

	filePaths = flag.Args()

	if len(filePaths) == 0 {
		fmt.Println("Please provide file paths as arguments.")
		return
	}

	results := findSymlinks(filePaths)

	// Write to the JSON file
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()

	prettyJSON, err := json.MarshalIndent(results, "", "    ")
	if err != nil {
		fmt.Printf("Error encoding JSON: %v\n", err)
		return
	}

	_, err = file.Write(prettyJSON)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
	}
}

func findSymlinks(filePaths []string) []FileLink {
	var results []FileLink
	var wg sync.WaitGroup
	outCh := make(chan FileLink, len(filePaths))
	skippedDirectories := strings.Split(skipDirs, ",")

	for _, filePath := range filePaths {
		wg.Add(1)
		go func(fp string) {
			defer wg.Done()
			var links []string
			var mu sync.Mutex // Mutex for link slice

			if debug {
				fmt.Printf("Starting goroutine for file: %s\n", fp)
			}

			fileInfo, err := os.Stat(fp)
			if err != nil {
				fmt.Printf("Error stating file: %v\n", err)
				return
			}

			fileSys := fileInfo.Sys()
			if fileSys == nil {
				return
			}

			_ = filepath.Walk("/", func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				for _, dir := range skippedDirectories {
					if strings.HasPrefix(path, dir) {
						return filepath.SkipDir
					}
				}
				if info.Mode()&os.ModeSymlink != 0 {
					realPath, _ := filepath.EvalSymlinks(path)
					if realPath == fp {
						mu.Lock()
						links = append(links, path)
						mu.Unlock()
					}
				}
				return nil
			})

			outCh <- FileLink{FilePath: fp, Symlinks: links}
		}(filePath)
	}

	go func() {
		wg.Wait()
		close(outCh)
	}()

	for link := range outCh {
		results = append(results, link)
	}

	return results
}
