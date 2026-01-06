package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"snap-memory-downloader/internal/app"
	"sync"
	"time"
)

func main() {
	defaultWorkers := runtime.NumCPU()
	inputPtr := flag.String("input", "", "Path to the HTML or JSON file. Defaults to memories_history.html or memory_history.json")
	outputPtr := flag.String("output", "./output", "Directory to save files")
	workersPtr := flag.Int("workers", defaultWorkers, "Number of concurrent downloads")
	skipImgPtr := flag.Bool("skip-image-overlay", false, "Skip overlays for image files")
	skipVidPtr := flag.Bool("skip-video-overlay", true, "Skip overlays for video files")
	keepArchPtr := flag.Bool("keep-archives", false, "Keep original ZIP files in overlays/archives/")
	dateFormatPtr := flag.String("date-format", "", "Custom date format for filenames (e.g., 'YYYYMMDD_HHMMSS' or 'YYMMDDTHHmmss'). Supported tokens: YYYY, YY, MM, DD, HH, hh, mm, ss")
	flag.Parse()

	cfg := app.Config{
		InputFile:        *inputPtr,
		OutputDir:        *outputPtr,
		Concurrency:      *workersPtr,
		SkipImageOverlay: *skipImgPtr,
		SkipVideoOverlay: *skipVidPtr,
		KeepArchives:     *keepArchPtr,
		DateFormat:       *dateFormatPtr,
	}

	// Handle default input file logic
	if cfg.InputFile == "" {
		if _, err := os.Stat("memories_history.html"); err == nil {
			cfg.InputFile = "memories_history.html"
		} else if _, err := os.Stat("memory_history.json"); err == nil {
			cfg.InputFile = "memory_history.json"
		} else {
			log.Fatalf("No input file specified and neither memories_history.html nor memory_history.json found.")
		}
	}

	content, err := os.ReadFile(cfg.InputFile)
	if err != nil {
		log.Fatalf("Error reading input file '%s': %v", cfg.InputFile, err)
	}

	var memories []app.MemoryItem
	ext := filepath.Ext(cfg.InputFile)
	switch ext {
	case ".html":
		memories = app.ParseHTML(string(content))
	case ".json":
		memories, err = app.ParseJSON(content)
		if err != nil {
			log.Fatalf("Error parsing JSON input file '%s': %v", cfg.InputFile, err)
		}
	default:
		log.Fatalf("Unsupported input file extension: %s. Only .html and .json are supported.", ext)
	}

	total := len(memories)
	if total == 0 {
		fmt.Println("No memories found.")
		return
	}

	fmt.Printf("Found %d memories. Using %d workers.\n", total, cfg.Concurrency)

	var wg sync.WaitGroup
	jobs := make(chan app.MemoryItem, total)
	progressChan := make(chan int, total)

	for w := 1; w <= cfg.Concurrency; w++ {
		wg.Add(1)
		go worker(jobs, progressChan, &wg, cfg)
	}

	for _, m := range memories {
		jobs <- m
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(progressChan)
	}()

	// Progress Monitoring
	completed := 0
	startTime := time.Now()
	for range progressChan {
		completed++
		app.PrintProgress(completed, total, startTime)
	}

	fmt.Println("\nTask finished.")
}

func worker(jobs <-chan app.MemoryItem, progress chan<- int, wg *sync.WaitGroup, cfg app.Config) {
	defer wg.Done()
	for item := range jobs {
		app.ProcessItem(item, cfg)
		progress <- 1
	}
}
