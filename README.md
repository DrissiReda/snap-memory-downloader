### Snapchat Memories Downloader

This tool automates archiving your Snapchat export with proper metadata. It matches GPS coordinates and timestamps to your media, merges overlays (filters and captions) onto images, and organizes everything into a clean directory structure.

---

### Prerequisites

You need the following tools in your system PATH but it only works in Linux:

| Tool | Purpose | Linux |
| --- | --- | --- |
| **FFmpeg** | Video merging | [ffmpeg.org](https://ffmpeg.org/download.html) |

---

### Building from Source

To build the application, ensure you have Go installed (version 1.24 or newer) and `make` is available on your system. Navigate to the project root and execute:

**Linux & macOS**

```bash
make linux
```

**Windows**

```powershell
make windows
```

---

### Usage

Place the binary in the same folder as your `memories_history.html` or `memory_history.json` and run:

```bash
./snap-memory-downloader -input memories_history.html -output ./MyMemories
# Or for JSON input
./snap-memory-downloader -input memory_history.json -output ./MyMemories
```

**Example Console Output**

```
Found 100 memories. Using 8 workers.
[==============---------------] 50/100 ETA: 10s
Task finished.
```

### Output

The output directory will have the following structure:

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
    │   └── ...
    └── videos/
        └── ...
```

**Arguments**

* `-input`: Path to the input file, which can be either an HTML (`.html`) or JSON (`.json`) file. If this flag is omitted, the program will first look for `memories_history.html`. If that's not found, it will then look for `memory_history.json`.
* `-output`: Directory for saved files (Default: `./output`)
* `-workers`: Concurrent downloads (Default: CPU core count)
* `-skip-video-overlay`: If set, videos are kept without overlays (default on Windows)
* `-keep-archives`: Saves original ZIP files in `overlays/archives/`

**Notes**

* For **Windows users**, just place `memories_history.html` in the same directory as the binary. Video overlay processing is disabled by default.
* For **Linux/macOS users**, you can enable video overlays with the `-skip-video-overlay=false` flag to merge filters and captions onto videos.

