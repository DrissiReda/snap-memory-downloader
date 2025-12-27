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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	"github.com/dsoprea/go-jpeg-image-structure/v2"
	"image/draw"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

type Config struct {
	InputFile        string
	OutputDir        string
	Concurrency      int
	SkipVideoOverlay bool
	KeepArchives     bool
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
	skipVidPtr := flag.Bool("skip-video-overlay", true, "Ignore overlays for video files")
	keepArchPtr := flag.Bool("keep-archives", false, "Keep original ZIP files in overlays/archives/")
	flag.Parse()

	config := Config{
		InputFile:        *inputPtr,
		OutputDir:        *outputPtr,
		Concurrency:      *workersPtr,
		SkipVideoOverlay: *skipVidPtr,
		KeepArchives:     *keepArchPtr,
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

	// Progress Monitoring
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
	fileBase := fmt.Sprintf("%s %s", item.Type, item.Date.Format("02-Jan-2006 15-04-05"))
	fileName := fileBase + item.Extension

	resp, err := http.Get(item.URL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var finalPath string
	if len(data) > 4 && bytes.HasPrefix(data, []byte("PK\x03\x04")) {
		overlayTypeDir := "images"
		if item.Extension == ".mp4" {
			overlayTypeDir = "videos"
		}

		// Keep Archive logic
		if config.KeepArchives {
			archFolder := filepath.Join(config.OutputDir, "overlays", "archives", year, month)
			os.MkdirAll(archFolder, os.ModePerm)
			os.WriteFile(filepath.Join(archFolder, fileBase+".zip"), data, 0644)
		}

		subFolder := filepath.Join(config.OutputDir, "overlays", overlayTypeDir, year, month)
		os.MkdirAll(subFolder, os.ModePerm)
		finalPath = filepath.Join(subFolder, fileName)
		handleZip(data, finalPath, item, config)
	} else {
		subFolder := filepath.Join(config.OutputDir, year, month)
		os.MkdirAll(subFolder, os.ModePerm)
		finalPath = filepath.Join(subFolder, fileName)
		os.WriteFile(finalPath, data, 0644)
	}

	// Native EXIF for JPG, Exiftool fallback for MP4
	if item.Extension == ".jpg" {
		lat, _ := strconv.ParseFloat(item.Latitude, 64)
		lon, _ := strconv.ParseFloat(item.Longitude, 64)
		_ = updateNativeExif(finalPath, lat, lon, item.Date)
	} else {
		runExifTool(finalPath, item)
	}
}

// --- Native Metadata Logic ---

func updateNativeExif(path string, lat, lon float64, dateTime time.Time) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	jmp := jpegstructure.NewJpegMediaParser()
	intfc, err := jmp.ParseBytes(data)
	if err != nil {
		return err
	}

	sl := intfc.(*jpegstructure.SegmentList)
	rootIb, err := sl.ConstructExifBuilder()
	if err != nil {
		return err
	}

	ifdIb, _ := exif.GetOrCreateIbFromRootIb(rootIb, "IFD0")
	exifIb, _ := exif.GetOrCreateIbFromRootIb(rootIb, "IFD0/Exif")
	gpsIb, _ := exif.GetOrCreateIbFromRootIb(rootIb, "IFD0/GPSInfo")

	dtStr := dateTime.Format("2006:01:02 15:04:05")
	_ = ifdIb.SetStandardWithName("DateTime", dtStr)
	_ = exifIb.SetStandardWithName("DateTimeOriginal", dtStr)
	_ = exifIb.SetStandardWithName("CreateDate", dtStr)

	if lat != 0 || lon != 0 {
		latRef, lonRef := "N", "E"
		if lat < 0 { latRef, lat = "S", -lat }
		if lon < 0 { lonRef, lon = "W", -lon }

		_ = gpsIb.SetStandardWithName("GPSLatitudeRef", latRef)
		_ = gpsIb.SetStandardWithName("GPSLongitudeRef", lonRef)
		_ = gpsIb.SetStandardWithName("GPSLatitude", decimalToRationals(lat))
		_ = gpsIb.SetStandardWithName("GPSLongitude", decimalToRationals(lon))
	}

	_ = sl.SetExif(rootIb)
	f, _ := os.Create(path)
	defer f.Close()
	_ = sl.Write(f)
	return os.Chtimes(path, dateTime, dateTime)
}

func decimalToRationals(decimal float64) []exifcommon.Rational {
	degrees := int(decimal)
	minutesFloat := (decimal - float64(degrees)) * 60
	minutes := int(minutesFloat)
	seconds := (minutesFloat - float64(minutes)) * 60
	return []exifcommon.Rational{
		{Numerator: uint32(degrees), Denominator: 1},
		{Numerator: uint32(minutes), Denominator: 1},
		{Numerator: uint32(seconds * 1000), Denominator: 1000},
	}
}

// --- Helper Functions (Zip, Video, Parsing) ---

func handleZip(data []byte, targetPath string, item MemoryItem, config Config) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil { return }
	var baseData, overlayData []byte
	var bName, oName string
	for _, file := range reader.File {
		buf := new(bytes.Buffer)
		f, _ := file.Open()
		io.Copy(buf, f)
		f.Close()
		if strings.Contains(file.Name, "-overlay") {
			overlayData, oName = buf.Bytes(), file.Name
		} else if strings.Contains(file.Name, "-main") {
			baseData, bName = buf.Bytes(), file.Name
		}
	}
	if baseData == nil { return }
	if item.Extension == ".jpg" && overlayData != nil {
		mergeImages(baseData, overlayData, targetPath)
	} else if item.Extension == ".mp4" && overlayData != nil && !config.SkipVideoOverlay {
		mergeVideos(baseData, overlayData, bName, oName, targetPath)
	} else {
		os.WriteFile(targetPath, baseData, 0644)
	}
}

func mergeImages(bgData, ovData []byte, outPath string) {
	bgImg, _, _ := image.Decode(bytes.NewReader(bgData))
	ovImg, _, _ := image.Decode(bytes.NewReader(ovData))
	if bgImg == nil || ovImg == nil { return }
	bounds := bgImg.Bounds()
	final := image.NewRGBA(bounds)
	draw.Draw(final, bounds, bgImg, image.Point{}, draw.Src)
	resizedOv := image.NewRGBA(bounds)
	xdraw.BiLinear.Scale(resizedOv, bounds, ovImg, ovImg.Bounds(), xdraw.Over, nil)
	draw.Draw(final, bounds, resizedOv, image.Point{}, draw.Over)
	f, _ := os.Create(outPath)
	defer f.Close()
	jpeg.Encode(f, final, &jpeg.Options{Quality: 90})
}

func mergeVideos(bgData, ovData []byte, bName, oName, outPath string) {
	tmpDir := os.TempDir()
	bTmp, oTmp := filepath.Join(tmpDir, bName), filepath.Join(tmpDir, oName)
	os.WriteFile(bTmp, bgData, 0644); os.WriteFile(oTmp, ovData, 0644)
	defer os.Remove(bTmp); defer os.Remove(oTmp)
	w, h := getVideoDimensions(bTmp)
	if w == "" { w, h = "540", "960" }
	filter := fmt.Sprintf("[1:v]scale=iw*%s/iw:ih*%s/ih[ovr];[0:v][ovr]overlay=0:0", w, h)
	exec.Command("ffmpeg", "-i", bTmp, "-i", oTmp, "-filter_complex", filter, "-pix_fmt", "yuv420p", "-c:a", "copy", outPath, "-y").Run()
}

func getVideoDimensions(path string) (string, string) {
	out, err := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "csv=s=x:p=0", path).Output()
	if err != nil { return "", "" }
	parts := strings.Split(strings.TrimSpace(string(out)), "x")
	if len(parts) == 2 { return parts[0], parts[1] }
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
		dateStr := strings.Replace(stripTags(cols[0][1]), " UTC", "", 1)
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
		items = append(items, MemoryItem{Date: t, Type: strings.TrimSpace(mType), Latitude: lat, Longitude: lon, URL: urlMatch[1], Extension: ext})
	}
	return items
}

func runExifTool(filePath string, item MemoryItem) {
	exifTime := item.Date.Format("2006:01:02 15:04:05")
	args := []string{"-overwrite_original", "-DateTimeOriginal=" + exifTime, "-CreateDate=" + exifTime, "-ModifyDate=" + exifTime, "-FileModifyDate=" + exifTime, "-MediaCreateDate=" + exifTime, "-MediaModifyDate=" + exifTime, "-SubSecCreateDate=" + exifTime}
	if item.Latitude != "" {
		args = append(args, "-GPSLatitude="+item.Latitude, "-GPSLongitude="+item.Longitude, "-GPSLatitudeRef<GPSLatitude", "-GPSLongitudeRef<GPSLongitude")
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
	
	elapsed := time.Since(start)
	// Calculate dynamic ETA based on speed
	itemsPerSecond := float64(current) / elapsed.Seconds()
	remainingItems := float64(total - current)
	var eta time.Duration
	if itemsPerSecond > 0 {
		eta = time.Duration(remainingItems/itemsPerSecond) * time.Second
	}
	
	fmt.Printf("\r[%s] %d/%d ETA: %s ", bar, current, total, eta.Round(time.Second).String())
}