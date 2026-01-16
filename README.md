# Snapchat Memories Downloader

Archive your Snapchat memories with proper metadata, GPS coordinates, timestamps, and overlays.


---

## Download

| Platform | Binary |
| --- | --- |
| **Windows** | [snap-memory-downloader.exe](https://github.com/DrissiReda/snap-memory-downloader/releases/latest/download/snap-memory-downloader.exe) |
| **macOS** | [snap-memory-downloader-darwin](https://github.com/DrissiReda/snap-memory-downloader/releases/latest/download/snap-memory-downloader-darwin) |
| **Linux** | [snap-memory-downloader](https://github.com/DrissiReda/snap-memory-downloader/releases/latest/download/snap-memory-downloader) |

---

## Usage

Run the executable to open the GUI:

- Drag & drop input files
- Configure parallel workers
- Custom date formats
- Toggle overlays
- Real-time progress
- Detailed logging

Here is what it looks like

<img width="806" height="646" alt="image" src="https://github.com/user-attachments/assets/0757325e-444f-41ba-8d9f-354c578dce62" />

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
