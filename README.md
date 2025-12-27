### Snapchat Memories Downloader

This tool automates archiving your Snapchat export with proper metadata. It matches GPS coordinates and timestamps to your media, merges overlays (filters and captions) onto images, and organizes everything into a clean directory structure.

---

### Prerequisites

You need the following tools in your system PATH:

| Tool | Purpose | Linux | Windows |
| --- | --- | --- | --- |
| **FFmpeg** | Video merging | `sudo apt install ffmpeg` | [ffmpeg.org](https://ffmpeg.org/download.html) |

---

### Building from Source

**Linux & macOS**

```bash
go get github.com/dsoprea/go-exif/v3
go get github.com/dsoprea/go-jpeg-image-structure/v2
go get golang.org/x/image/draw
go build -o snap-memory-downloader main.go
````

**Windows**

```powershell
go build -o snap-memory-downloader.exe main.go
```

---

### Usage

Place the binary in the same folder as your `memories_history.html` and run:

```bash
./snap-memory-downloader -input memories_history.html -output ./MyMemories
```

**Arguments**

* `-input`: Path to the HTML file (Default: `memories_history.html`)
* `-output`: Directory for saved files (Default: `./output`)
* `-workers`: Concurrent downloads (Default: CPU core count)
* `-skip-video-overlay`: If set, videos are kept without overlays (default on Windows)
* `-keep-archives`: Saves original ZIP files in `overlays/archives/`

**Notes**

* For **Windows users**, just place `memories_history.html` in the same directory as the binary. Video overlay processing is disabled by default.
* For **Linux/macOS users**, you can enable video overlays with the `-skip-video-overlay=false` flag to merge filters and captions onto videos.

