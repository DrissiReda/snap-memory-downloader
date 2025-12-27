package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// mergeVideos merges a background video with an overlay using ffmpeg.
func mergeVideos(bgData, ovData []byte, bName, oName, outPath string) {
	tmpDir := os.TempDir()
	bTmp, oTmp := filepath.Join(tmpDir, bName), filepath.Join(tmpDir, oName)
	os.WriteFile(bTmp, bgData, 0644)
	os.WriteFile(oTmp, ovData, 0644)
	defer os.Remove(bTmp)
	defer os.Remove(oTmp)
	w, h := getVideoDimensions(bTmp)
	if w == "" {
		w, h = "540", "960"
	}
	filter := fmt.Sprintf("[1:v]scale=iw*%s/iw:ih*%s/ih[ovr];[0:v][ovr]overlay=0:0", w, h)
	exec.Command("ffmpeg", "-i", bTmp, "-i", oTmp, "-filter_complex", filter, "-pix_fmt", "yuv420p", "-c:a", "copy", outPath, "-y").Run()
}

// getVideoDimensions extracts video width and height using ffprobe.
func getVideoDimensions(path string) (string, string) {
	out, err := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "csv=s=x:p=0", path).Output()
	if err != nil {
		return "", ""
	}
	parts := strings.Split(strings.TrimSpace(string(out)), "x")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}
