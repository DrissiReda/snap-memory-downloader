package test

import (
	"net/http"
	"net/http/httptest"
	"snap-memory-downloader/internal/app"
	"testing"
	"time"
)

func TestParseHTML(t *testing.T) {
	html := `
		<table>
			<tbody>
				<tr>
					<td>2023-10-27 10:00:00 UTC</td>
					<td>Test Type</td>
					<td>34.052235, -118.243683</td>
					<td><a href="#" onclick="return downloadMemories('http://example.com/memory1.jpg')">Download</a></td>
				</tr>
				<tr>
					<td>2023-10-28 11:00:00 UTC</td>
					<td>Test Video</td>
					<td>40.712776, -74.005974</td>
					<td><a href="#" onclick="return downloadMemories('http://example.com/memory2.mp4')">Download</a></td>
				</tr>
			</tbody>
		</table>
	`

	expected := []app.MemoryItem{
		{
			Date:      time.Date(2023, 10, 27, 10, 0, 0, 0, time.UTC),
			Type:      "Test Type",
			Latitude:  "34.052235",
			Longitude: "-118.243683",
			URL:       "http://example.com/memory1.jpg",
			Extension: ".jpg",
		},
		{
			Date:      time.Date(2023, 10, 28, 11, 0, 0, 0, time.UTC),
			Type:      "Test Video",
			Latitude:  "40.712776",
			Longitude: "-74.005974",
			URL:       "http://example.com/memory2.mp4",
			Extension: ".mp4",
		},
	}

	items := app.ParseHTML(html)

	if len(items) != len(expected) {
		t.Fatalf("Expected %d items, but got %d", len(expected), len(items))
	}

	for i, item := range items {
		if !item.Date.Equal(expected[i].Date) {
			t.Errorf("Item %d: Expected date %v, but got %v", i, expected[i].Date, item.Date)
		}
		if item.Type != expected[i].Type {
			t.Errorf("Item %d: Expected type %s, but got %s", i, expected[i].Type, item.Type)
		}
		if item.Latitude != expected[i].Latitude {
			t.Errorf("Item %d: Expected latitude %s, but got %s", i, expected[i].Latitude, item.Latitude)
		}
		if item.Longitude != expected[i].Longitude {
			t.Errorf("Item %d: Expected longitude %s, but got %s", i, expected[i].Longitude, item.Longitude)
		}
		if item.URL != expected[i].URL {
			t.Errorf("Item %d: Expected URL %s, but got %s", i, expected[i].URL, item.URL)
		}
		if item.Extension != expected[i].Extension {
			t.Errorf("Item %d: Expected extension %s, but got %s", i, expected[i].Extension, item.Extension)
		}
	}
}

func TestStripTags(t *testing.T) {
	input := "<p>Hello, <b>world</b>!</p>"
	expected := "Hello, world!"
	if app.StripTags(input) != expected {
		t.Errorf("Expected %s, but got %s", expected, app.StripTags(input))
	}
}

func TestDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test data"))
	}))
	defer server.Close()

	data, err := app.DownloadFile(server.URL)
	if err != nil {
		t.Fatalf("Expected no error, but got %v", err)
	}

	if string(data) != "test data" {
		t.Errorf("Expected 'test data', but got '%s'", string(data))
	}
}

func TestIsZip(t *testing.T) {
	zipData := []byte("PK\x03\x04\x14\x00\x00\x00\x08\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")
	if !app.IsZip(zipData) {
		t.Errorf("Expected true for zip data, but got false")
	}

	nonZipData := []byte("this is not a zip file")
	if app.IsZip(nonZipData) {
		t.Errorf("Expected false for non-zip data, but got true")
	}
}
