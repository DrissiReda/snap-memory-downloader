# Snapchat Memories Downloader

Archive your Snapchat memories with proper metadata, GPS coordinates, timestamps, and overlays.

---

## Download

| Platform | Binary |
| --- | --- |
| **Windows** | [snap-memory-downloader-windows-amd64.exe](https://github.com/yourusername/snap-memory-downloader/releases/latest/download/snap-memory-downloader-windows-amd64.exe) |
| **macOS** | [snap-memory-downloader-darwin-amd64](https://github.com/yourusername/snap-memory-downloader/releases/latest/download/snap-memory-downloader-darwin-amd64) |
| **Linux** | [snap-memory-downloader-linux-amd64](https://github.com/yourusername/snap-memory-downloader/releases/latest/download/snap-memory-downloader-linux-amd64) |

---

## Usage

Run the executable to open the GUI:

- Drag & drop input files
- Configure parallel workers
- Custom date formats
- Toggle overlays
- Real-time progress
- Detailed logging

### Requirements

FFmpeg (Linux/macOS only, for video overlays): [ffmpeg.org](https://ffmpeg.org/download.html)

If you want video overlays to work in windows you'll have to have `ffmpeg` and `ffprobe` in your `PATH` 

---

## Output Structure

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

## Notes

- Windows: Video overlays unavailable
- Linux/macOS: Full overlay support (requires FFmpeg)
- Auto-detects memories_history.html or memory_history.json
- EXIF metadata applied automatically

---

## License

MIT License