package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"image/draw" // Standard library for drawing operations
	xdraw "golang.org/x/image/draw" // Aliased for high-quality scaling
	_ "golang.org/x/image/webp"
)

type Config struct {
	InputFile        string
	OutputDir        string
	Concurrency      int
	SkipVideoOverlay bool
}

type MemoryItem struct {
	Date      time.Time
	Type      string
	Latitude  string
	Longitude string
	URL       string
	Extension string
}

func main() {
	defaultWorkers := runtime.NumCPU()
	inputPtr := flag.String("input", "memories_history.html", "Path to the HTML file")
	outputPtr := flag.String("output", "./output", "Directory to save files")
	workersPtr := flag.Int("workers", defaultWorkers, "Number of concurrent downloads")
	skipVidPtr := flag.Bool("skip-video-overlay", false, "Ignore overlays for video files")
	flag.Parse()

	config := Config{
		InputFile:        *inputPtr,
		OutputDir:        *outputPtr,
		Concurrency:      *workersPtr,
		SkipVideoOverlay: *skipVidPtr,
	}

	content, err := os.ReadFile(config.InputFile)
	if err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	memories := parseHTML(string(content))
	total := len(memories)
	if total == 0 {
		fmt.Println("No memories found.")
		return
	}

	fmt.Printf("Found %d memories. Using %d workers.\n", total, config.Concurrency)

	var wg sync.WaitGroup
	jobs := make(chan MemoryItem, total)
	progressChan := make(chan int, total)

	for w := 1; w <= config.Concurrency; w++ {
		wg.Add(1)
		go worker(jobs, progressChan, &wg, config)
	}

	for _, m := range memories {
		jobs <- m
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(progressChan)
	}()

	completed := 0
	startTime := time.Now()
	for range progressChan {
		completed++
		printProgress(completed, total, startTime)
	}
	fmt.Println("\nTask finished.")
}

func worker(jobs <-chan MemoryItem, progress chan<- int, wg *sync.WaitGroup, config Config) {
	defer wg.Done()
	for item := range jobs {
		processItem(item, config)
		progress <- 1
	}
}

func processItem(item MemoryItem, config Config) {
	year, month := item.Date.Format("2006"), item.Date.Format("01")
	subFolder := filepath.Join(config.OutputDir, year, month)
	os.MkdirAll(subFolder, os.ModePerm)

	fileName := fmt.Sprintf("%s %s%s", item.Type, item.Date.Format("02-Jan-2006 15-04-05"), item.Extension)
	fullPath := filepath.Join(subFolder, fileName)

	resp, err := http.Get(item.URL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if len(data) > 4 && bytes.HasPrefix(data, []byte("PK\x03\x04")) {
		handleZip(data, fullPath, item, config)
	} else {
		os.WriteFile(fullPath, data, 0644)
	}

	runExifTool(fullPath, item)
}

func handleZip(data []byte, targetPath string, item MemoryItem, config Config) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return
	}

	var baseData, overlayData []byte
	var baseName, overlayName string

	for _, file := range reader.File {
		buf := new(bytes.Buffer)
		f, _ := file.Open()
		io.Copy(buf, f)
		f.Close()

		if strings.Contains(file.Name, "-overlay") {
			overlayData = buf.Bytes()
			overlayName = file.Name
		} else if strings.Contains(file.Name, "-main") {
			baseData = buf.Bytes()
			baseName = file.Name
		}
	}

	if baseData == nil {
		return
	}

	if item.Extension == ".jpg" {
		if overlayData != nil {
			mergeImages(baseData, overlayData, targetPath)
		} else {
			os.WriteFile(targetPath, baseData, 0644)
		}
		return
	}

	if item.Extension == ".mp4" {
		if overlayData != nil && !config.SkipVideoOverlay {
			mergeVideos(baseData, overlayData, baseName, overlayName, targetPath)
		} else {
			os.WriteFile(targetPath, baseData, 0644)
		}
	}
}

func mergeImages(bgData, ovData []byte, outPath string) {
	bgImg, _, _ := image.Decode(bytes.NewReader(bgData))
	ovImg, _, _ := image.Decode(bytes.NewReader(ovData))
	if bgImg == nil || ovImg == nil {
		os.WriteFile(outPath, bgData, 0644)
		return
	}

	bounds := bgImg.Bounds()
	final := image.NewRGBA(bounds)

	// Draw background using standard library
	draw.Draw(final, bounds, bgImg, image.Point{}, draw.Src)

	// Create a temporary buffer for the resized overlay
	resizedOv := image.NewRGBA(bounds)
	// Use xdraw (aliased golang.org/x/image/draw) for scaling
	xdraw.BiLinear.Scale(resizedOv, bounds, ovImg, ovImg.Bounds(), xdraw.Over, nil)

	// Draw the resized overlay onto the final image
	draw.Draw(final, bounds, resizedOv, image.Point{}, draw.Over)

	f, _ := os.Create(outPath)
	defer f.Close()
	jpeg.Encode(f, final, &jpeg.Options{Quality: 90})
}

func mergeVideos(bgData, ovData []byte, bName, oName, outPath string) {
	tmpDir := os.TempDir()
	bTmp := filepath.Join(tmpDir, bName)
	oTmp := filepath.Join(tmpDir, oName)
	os.WriteFile(bTmp, bgData, 0644)
	os.WriteFile(oTmp, ovData, 0644)
	defer os.Remove(bTmp)
	defer os.Remove(oTmp)

	width, height := getVideoDimensions(bTmp)
	if width == "" {
		width, height = "540", "960"
	}

	filter := fmt.Sprintf("[1:v]scale=iw*%s/iw:ih*%s/ih[ovr];[0:v][ovr]overlay=0:0", width, height)
	
	cmd := exec.Command("ffmpeg", "-i", bTmp, "-i", oTmp,
		"-filter_complex", filter,
		"-pix_fmt", "yuv420p", "-c:a", "copy",
		outPath, "-y")

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		fmt.Printf("\nFFmpeg Error on %s: %v\nLogs: %s", outPath, err, errBuf.String())
	}
}

func getVideoDimensions(path string) (string, string) {
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0",
		"-show_entries", "stream=width,height", "-of", "csv=s=x:p=0", path)
	out, err := cmd.Output()
	if err != nil {
		return "", ""
	}
	parts := strings.Split(strings.TrimSpace(string(out)), "x")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}

func parseHTML(html string) []MemoryItem {
	var items []MemoryItem
	rowRegex := regexp.MustCompile(`(?s)<tr>(.*?)</tr>`)
	colRegex := regexp.MustCompile(`(?s)<td>(.*?)</td>`)
	gpsRegex := regexp.MustCompile(`([-+]?\d*\.\d+|\d+)`)
	urlRegex := regexp.MustCompile(`downloadMemories\('([^']+)'`)

	rows := rowRegex.FindAllStringSubmatch(html, -1)
	for _, rowMatch := range rows {
		cols := colRegex.FindAllStringSubmatch(rowMatch[1], -1)
		if len(cols) < 4 { continue }

		dateStr := stripTags(cols[0][1])
		dateStr = strings.Replace(dateStr, " UTC", "", 1)
		t, err := time.Parse("2006-01-02 15:04:05", dateStr)
		if err != nil { continue }

		mType := stripTags(cols[1][1])
		ext := ".jpg"
		if strings.Contains(mType, "Video") { ext = ".mp4" }

		gpsStr := stripTags(cols[2][1])
		gps := gpsRegex.FindAllString(gpsStr, -1)
		lat, lon := "", ""
		if len(gps) >= 2 { lat, lon = gps[0], gps[1] }

		urlMatch := urlRegex.FindStringSubmatch(cols[3][1])
		if len(urlMatch) < 2 { continue }

		items = append(items, MemoryItem{
			Date: t, Type: strings.TrimSpace(mType),
			Latitude: lat, Longitude: lon,
			URL: urlMatch[1], Extension: ext,
		})
	}
	return items
}

func runExifTool(filePath string, item MemoryItem) {
	exifTime := item.Date.Format("2006:01:02 15:04:05")
	args := []string{"-overwrite_original",
		"-DateTimeOriginal=" + exifTime, "-CreateDate=" + exifTime, "-ModifyDate=" + exifTime,
		"-FileModifyDate=" + exifTime, "-MediaCreateDate=" + exifTime,
		"-MediaModifyDate=" + exifTime, "-SubSecCreateDate=" + exifTime,
	}
	if item.Latitude != "" {
		args = append(args, "-GPSLatitude="+item.Latitude, "-GPSLongitude="+item.Longitude,
			"-GPSLatitudeRef<GPSLatitude", "-GPSLongitudeRef<GPSLongitude")
	}
	args = append(args, filePath)
	exec.Command("exiftool", args...).Run()
}

func stripTags(input string) string {
	return regexp.MustCompile(`<[^>]*>`).ReplaceAllString(input, "")
}

func printProgress(current, total int, start time.Time) {
	percent := float64(current) / float64(total)
	bar := strings.Repeat("=", int(percent*30)) + strings.Repeat("-", 30-int(percent*30))
	eta := time.Duration(float64(time.Since(start)) / float64(current) * float64(total-current)).Round(time.Second)
	fmt.Printf("\r[%s] %d/%d ETA: %v ", bar, current, total, eta)
}