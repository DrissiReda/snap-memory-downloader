# Snapchat Memories Downloader

This project helps you better download your memories.

It tries to match Date and GPS as metadata to your images and videos.

It also unzips the overlayed zip archives, then adds back the overlay to the images and videos.

## Requirements (Linux)

- Exiftool
- FFmpeg


## Usage

Use the binaries in release using whatever flags you like:

```bash
  -input string
        Path to the HTML file (default "memories_history.html")
  -output string
        Directory to save files (default "./output")
  -skip-video-overlay
        Ignore overlays for video files
  -workers int
        Number of concurrent downloads (default 32)
```



## Coming soon

I don't have a windows machine yet so I need time to add these features:

- Metadata modification.
- Overlays for windows


This repo was made in less than 30 minutes so please forgive how messy it is, especially the README.md

Contributions welcome for windows support