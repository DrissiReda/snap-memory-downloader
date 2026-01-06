package app

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Config holds the application's configuration settings.
type Config struct {
	InputFile        string
	OutputDir        string
	Concurrency      int
	SkipImageOverlay bool
	SkipVideoOverlay bool
	KeepArchives     bool
	DateFormat       string
}

// MemoryItem represents a single memory item extracted from the HTML file.
type MemoryItem struct {
	Date      time.Time
	Type      string
	Latitude  string
	Longitude string
	URL       string
	Extension string
}

// jsonMemoryItem is a helper struct for unmarshaling JSON input.
type jsonMemoryItem struct {
	Date             string `json:"Date"`
	MediaType        string `json:"Media Type"`
	Location         string `json:"Location"`
	DownloadLink     string `json:"Download Link"`
	MediaDownloadUrl string `json:"Media Download Url"`
}

// jsonInput is a helper struct for unmarshaling the overall JSON structure.
type jsonInput struct {
	SavedMedia []jsonMemoryItem `json:"Saved Media"`
}

// ParseHTML extracts memory items from the HTML content.
func ParseHTML(html string) []MemoryItem {
	var items []MemoryItem
	rowRegex := regexp.MustCompile(`(?s)<tr>(.*?)</tr>`)
	colRegex := regexp.MustCompile(`(?s)<td>(.*?)</td>`)
	gpsRegex := regexp.MustCompile(`([-+]?\d*\.\d+|\d+)`)
	urlRegex := regexp.MustCompile(`downloadMemories\('([^']+)'`)
	rows := rowRegex.FindAllStringSubmatch(html, -1)
	for _, rowMatch := range rows {
		cols := colRegex.FindAllStringSubmatch(rowMatch[1], -1)
		if len(cols) < 4 {
			continue
		}
		dateStr := strings.Replace(StripTags(cols[0][1]), " UTC", "", 1)
		t, err := time.Parse("2006-01-02 15:04:05", dateStr)
		if err != nil {
			continue
		}
		mType := StripTags(cols[1][1])
		ext := ".jpg"
		if strings.Contains(mType, "Video") {
			ext = ".mp4"
		}
		gpsStr := StripTags(cols[2][1])
		gps := gpsRegex.FindAllString(gpsStr, -1)
		lat, lon := "", ""
		if len(gps) >= 2 {
			lat, lon = gps[0], gps[1]
		}
		urlMatch := urlRegex.FindStringSubmatch(cols[3][1])
		if len(urlMatch) < 2 {
			continue
		}
		items = append(items, MemoryItem{Date: t, Type: strings.TrimSpace(mType), Latitude: lat, Longitude: lon, URL: urlMatch[1], Extension: ext})
	}
	return items
}

// ParseJSON extracts memory items from JSON content.
func ParseJSON(jsonData []byte) ([]MemoryItem, error) {
	var input jsonInput
	if err := json.Unmarshal(jsonData, &input); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	var items []MemoryItem
	gpsRegex := regexp.MustCompile(`([-+]?\d*\.\d+|\d+)`)
	for _, jItem := range input.SavedMedia {
		t, err := time.Parse("2006-01-02 15:04:05 UTC", jItem.Date)
		if err != nil {
			return nil, fmt.Errorf("error parsing date '%s': %w", jItem.Date, err)
		}

		ext := ".jpg"
		if strings.Contains(jItem.MediaType, "Video") {
			ext = ".mp4"
		}

		gps := gpsRegex.FindAllString(jItem.Location, -1)
		lat, lon := "", ""
		if len(gps) >= 2 {
			lat, lon = gps[0], gps[1]
		}

		items = append(items, MemoryItem{
			Date:      t,
			Type:      strings.TrimSpace(jItem.MediaType),
			Latitude:  lat,
			Longitude: lon,
			URL:       jItem.MediaDownloadUrl,
			Extension: ext,
		})
	}
	return items, nil
}

// StripTags removes HTML tags from a string.
func StripTags(input string) string {
	return regexp.MustCompile(`<[^>]*>`).ReplaceAllString(input, "")
}

// DownloadFile downloads a file from the given URL and returns its content.
func DownloadFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// HandleZip processes a ZIP archive containing media and overlays.
func HandleZip(data []byte, targetPath string, item MemoryItem, config Config) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return
	}
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
	if baseData == nil {
		return
	}

	// Check skip flags based on media type
	skipOverlay := (item.Extension == ".jpg" && config.SkipImageOverlay) ||
		(item.Extension == ".mp4" && config.SkipVideoOverlay)

	if skipOverlay {
		os.WriteFile(targetPath, baseData, 0644)
		return
	}

	if item.Extension == ".jpg" && overlayData != nil {
		mergeImages(baseData, overlayData, targetPath)
	} else if item.Extension == ".mp4" && overlayData != nil {
		mergeVideos(baseData, overlayData, bName, oName, targetPath)
	} else {
		os.WriteFile(targetPath, baseData, 0644)
	}
}

// ProcessItem handles the downloading, processing, and saving of a single memory item.
func ProcessItem(item MemoryItem, config Config) {
	data, err := DownloadFile(item.URL)
	if err != nil {
		return
	}

	year, month := item.Date.Format("2006"), item.Date.Format("01")

	// Use custom date format if provided
	dateStr := item.Date.Format("02-Jan-2006 15-04-05")
	if config.DateFormat != "" {
		dateStr = FormatDateCustom(item.Date, config.DateFormat)
	}

	fileBase := fmt.Sprintf("%s %s", item.Type, dateStr)
	fileName := fileBase + item.Extension

	var finalPath string
	if IsZip(data) {
		finalPath = handleZippedItem(item, data, config, year, month, fileBase, fileName)
	} else {
		finalPath = handleRegularItem(item, data, config, year, month, fileName)
	}

	applyMetadata(finalPath, item)
}

// FormatDateCustom formats a time according to a custom format string.
func FormatDateCustom(t time.Time, format string) string {
	// Replace in order: longer patterns first to avoid conflicts
	replacements := []struct {
		pattern   string
		goPattern string
	}{
		{"YYYY", "2006"},
		{"YY", "06"},
		{"MM", "01"},
		{"DD", "02"},
		{"HH", "15"},
		{"hh", "03"},
		{"mm", "04"},
		{"ss", "05"},
		{"SS", "05"}, // Support both ss and SS for seconds
	}

	goFormat := format
	for _, r := range replacements {
		goFormat = strings.ReplaceAll(goFormat, r.pattern, r.goPattern)
	}

	return t.Format(goFormat)
}

// IsZip checks if the given data is a ZIP archive.
func IsZip(data []byte) bool {
	return len(data) > 4 && bytes.HasPrefix(data, []byte("PK\x03\x04"))
}

// handleZippedItem processes a memory item that is a ZIP archive.
func handleZippedItem(item MemoryItem, data []byte, config Config, year, month, fileBase, fileName string) string {
	overlayTypeDir := "images"
	if item.Extension == ".mp4" {
		overlayTypeDir = "videos"
	}

	if config.KeepArchives {
		archiveFolder := filepath.Join(config.OutputDir, "overlays", "archives", year, month)
		os.MkdirAll(archiveFolder, os.ModePerm)
		os.WriteFile(filepath.Join(archiveFolder, fileBase+".zip"), data, 0644)
	}

	subFolder := filepath.Join(config.OutputDir, "overlays", overlayTypeDir, year, month)
	os.MkdirAll(subFolder, os.ModePerm)
	finalPath := filepath.Join(subFolder, fileName)
	HandleZip(data, finalPath, item, config)
	return finalPath
}

// handleRegularItem processes a memory item that is not a ZIP archive.
func handleRegularItem(item MemoryItem, data []byte, config Config, year, month, fileName string) string {
	subFolder := filepath.Join(config.OutputDir, year, month)
	os.MkdirAll(subFolder, os.ModePerm)
	finalPath := filepath.Join(subFolder, fileName)
	os.WriteFile(finalPath, data, 0644)
	return finalPath
}

// applyMetadata applies EXIF data to the processed file.
func applyMetadata(path string, item MemoryItem) {
	if item.Extension == ".jpg" {
		lat, _ := strconv.ParseFloat(item.Latitude, 64)
		lon, _ := strconv.ParseFloat(item.Longitude, 64)
		_ = updateNativeExif(path, lat, lon, item.Date)
	}
}

// PrintProgress displays a progress bar in the console.
func PrintProgress(current, total int, start time.Time) {
	percent := float64(current) / float64(total)
	bar := strings.Repeat("=", int(percent*30)) + strings.Repeat("-", 30-int(percent*30))

	elapsed := time.Since(start)
	itemsPerSecond := float64(current) / elapsed.Seconds()
	remainingItems := float64(total - current)
	var eta time.Duration
	if itemsPerSecond > 0 {
		eta = time.Duration(remainingItems/itemsPerSecond) * time.Second
	}

	fmt.Printf("\r[%s] %d/%d ETA: %s ", bar, current, total, eta.Round(time.Second).String())
}
