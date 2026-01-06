### Snapchat Memories Downloader

This tool automates archiving your Snapchat export with proper metadata. It matches GPS coordinates and timestamps to your media, merges overlays (filters and captions) onto images, and organizes everything into a clean directory structure.

---

### Prerequisites

You need the following tools in your system PATH (Linux/macOS only for video overlays):

| Tool | Purpose | Download |
| --- | --- | --- |
| **FFmpeg** | Video overlay merging | [ffmpeg.org](https://ffmpeg.org/download.html) |

---

### Usage

**Download the Binary**

Download the appropriate binary for your system from releases:
- **macOS**: `snap-memory-downloader-darwin-amd64`
- **Linux**: `snap-memory-downloader-linux-amd64`
- **Windows**: `snap-memory-downloader-windows-amd64.exe`

**Running the Tool**

Place the binary in the same folder as your `memories_history.html` or `memory_history.json`.

**Windows**
```cmd
snap-memory-downloader-windows-amd64.exe -input memories_history.html -output ./MyMemories
```

**macOS**
```bash
chmod +x snap-memory-downloader-darwin-amd64
./snap-memory-downloader-darwin-amd64 -input memories_history.html -output ./MyMemories
```

**Linux**
```bash
chmod +x snap-memory-downloader-linux-amd64
./snap-memory-downloader-linux-amd64 -input memories_history.html -output ./MyMemories
```

**Quick Start (Windows)**

Double-click the `.exe` file to download everything locally with default settings. Images will have overlays, videos will be downloaded without overlays.

**Example Console Output**
```
Found 100 memories. Using 8 workers.
[==============---------------] 50/100 ETA: 10s
Task finished.
```

---

### Command-Line Arguments

| Flag | Description | Default |
| --- | --- | --- |
| `-input` | Path to HTML or JSON file | Looks for `memories_history.html`, then `memory_history.json` |
| `-output` | Output directory | `./output` |
| `-workers` | Concurrent downloads | CPU core count |
| `-skip-image-overlay` | Skip overlays for images | `false` |
| `-skip-video-overlay` | Skip overlays for videos | `true` |
| `-keep-archives` | Save original ZIP files | `false` |
| `-date-format` | Custom filename date format | `DD-MMM-YYYY HH-mm-ss` |

**Date Format Examples**

The `-date-format` flag accepts custom patterns:

```bash
# Format: YYYYMMDD_HHMMSS
./snap-memory-downloader -date-format "YYYYMMDD_HHMMSS"
# Output: Photo 20230115_143005.jpg

# Format: YYMMDD-HHmmss
./snap-memory-downloader -date-format "YYMMDD-HHmmss"
# Output: Photo 230115-143005.jpg

# Format: YYYY-MM-DD at HH:mm:ss
./snap-memory-downloader -date-format "YYYY-MM-DD at HH:mm:ss"
# Output: Photo 2023-01-15 at 14:30:05.jpg
```

**Supported Tokens**
- `YYYY` - 4-digit year (2023)
- `YY` - 2-digit year (23)
- `MM` - 2-digit month (01-12)
- `DD` - 2-digit day (01-31)
- `HH` - 2-digit hour 24h format (00-23)
- `hh` - 2-digit hour 12h format (01-12)
- `mm` - 2-digit minute (00-59)
- `ss` - 2-digit second (00-59)

---

### Building from Source

Ensure you have Go 1.24+ and `make` installed.

**Linux & macOS**
```bash
make linux
```

**Windows**
```powershell
make windows
```

---

### Output Structure

```
MyMemories/
├── 2023/
│   ├── 01/
│   │   ├── Photo 01-Jan-2023 15-04-05.jpg
│   │   └── Video 02-Jan-2023 16-30-00.mp4
│   └── 02/
│       └── ...
└── overlays/
    ├── images/
    │   └── 2023/
    │       └── 01/
    │           └── Photo 01-Jan-2023 15-04-05.jpg
    ├── videos/
    │   └── 2023/
    │       └── 01/
    │           └── Video 02-Jan-2023 16-30-00.mp4
    └── archives/  (if -keep-archives is set)
        └── 2023/
            └── 01/
                └── Photo 01-Jan-2023 15-04-05.zip
```

---

### Notes

- **Windows**: Video overlay processing is unavailable. Use `-skip-image-overlay` to also skip image overlays.
- **Linux/macOS**: Full overlay support for both images and videos (requires FFmpeg for videos). Set `-skip-video-overlay=false` to enable video overlays.
- If no input file is specified, the tool searches for `memories_history.html` first, then `memory_history.json`.
- EXIF metadata (GPS coordinates, timestamps) is automatically applied to images.