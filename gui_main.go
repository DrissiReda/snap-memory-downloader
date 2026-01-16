package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"runtime"
	"snap-memory-downloader/internal/app"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type GuiApp struct {
	window         fyne.Window
	inputFile      *widget.Entry
	outputDir      *widget.Entry
	workers        *widget.Entry
	skipImageCheck *widget.Check
	skipVideoCheck *widget.Check
	keepArchCheck  *widget.Check
	dateFormat     *widget.Entry
	debugCheck     *widget.Check
	progressBar    *widget.ProgressBar
	statusLabel    *widget.Label
	logOutput      *widget.Entry
	startButton    *widget.Button
	tabs           *container.AppTabs
	isProcessing   bool
}

func main() {
	a := fyneapp.NewWithID("com.snapmemory.downloader")
	a.Settings().SetTheme(&modernTheme{})

	guiApp := &GuiApp{}
	guiApp.window = a.NewWindow("Snap Memory Downloader")
	guiApp.window.Resize(fyne.NewSize(800, 600))

	guiApp.setupUI()
	guiApp.window.ShowAndRun()
}

func (g *GuiApp) setupUI() {
	// Set up window-level drop handling
	g.window.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		if len(uris) > 0 {
			// Use the first dropped file
			g.inputFile.SetText(uris[0].Path())
		}
	})

	// Main configuration tab
	configContent := g.createConfigTab()

	// Logs tab
	logsContent := g.createLogsTab()

	// Create tabs
	g.tabs = container.NewAppTabs(
		container.NewTabItem("Configuration", configContent),
		container.NewTabItem("Logs", logsContent),
	)

	g.window.SetContent(g.tabs)
}

func (g *GuiApp) createConfigTab() fyne.CanvasObject {
	// Helper for small labels
	smallLabel := func(text string) fyne.CanvasObject {
		return widget.NewLabel(text)
	}

	// Input file section with drag-and-drop
	g.inputFile = widget.NewEntry()
	g.inputFile.SetPlaceHolder("Drop HTML/JSON file here...")
	g.inputFile.OnChanged = func(s string) {} // Compact

	inputBrowse := widget.NewButton("...", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			g.inputFile.SetText(reader.URI().Path())
			reader.Close()
		}, g.window)
	})
	inputBrowse.Importance = widget.LowImportance

	// Output directory
	g.outputDir = widget.NewEntry()
	g.outputDir.SetText("./output")

	outputBrowse := widget.NewButton("...", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			g.outputDir.SetText(uri.Path())
		}, g.window)
	})
	outputBrowse.Importance = widget.LowImportance

	// Workers
	g.workers = widget.NewEntry()
	g.workers.SetText(fmt.Sprintf("%d", runtime.NumCPU()))

	// Date format
	g.dateFormat = widget.NewEntry()
	g.dateFormat.SetPlaceHolder("YYYYMMDD_HHMMSS")

	// Input row with label
	inputRow := container.NewBorder(nil, nil, nil, inputBrowse, g.inputFile)
	inputSection := container.NewVBox(smallLabel("Input File:"), inputRow)

	// Output row with label
	outputRow := container.NewBorder(nil, nil, nil, outputBrowse, g.outputDir)
	outputSection := container.NewVBox(smallLabel("Output Directory:"), outputRow)

	// Settings row (Workers and Date Format)
	workersSection := container.NewVBox(smallLabel("Workers:"), g.workers)
	dateSection := container.NewVBox(smallLabel("Date Format:"), g.dateFormat)
	settingsRow := container.NewGridWithColumns(2, workersSection, dateSection)

	// Options
	g.skipImageCheck = widget.NewCheck("Image overlays", func(bool) {})
	g.skipImageCheck.SetChecked(false)

	g.skipVideoCheck = widget.NewCheck("Video overlays", func(bool) {})
	g.skipVideoCheck.SetChecked(true)

	g.keepArchCheck = widget.NewCheck("Archive files", func(bool) {})
	g.keepArchCheck.SetChecked(false)

	g.debugCheck = widget.NewCheck("Debug logging", func(bool) {})
	g.debugCheck.SetChecked(false)

	optionsRow := container.NewGridWithColumns(2,
		g.skipImageCheck,
		g.skipVideoCheck,
		g.keepArchCheck,
		g.debugCheck,
	)

	// Progress section
	g.progressBar = widget.NewProgressBar()
	g.statusLabel = widget.NewLabel("Ready to start")

	// Start button
	g.startButton = widget.NewButtonWithIcon("Start", theme.DownloadIcon(), func() {
		g.startProcessing()
	})
	g.startButton.Importance = widget.HighImportance

	// Progress and button on same line
	progressContainer := container.NewBorder(nil, nil, nil, g.startButton, g.progressBar)
	progressSection := container.NewVBox(
		g.statusLabel,
		progressContainer,
	)

	// Section header
	createHeader := func(title string) fyne.CanvasObject {
		return widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	}

	// Main layout with clear section separation
	content := container.NewVBox(
		createHeader("Input & Output"),
		inputSection,
		outputSection,
		layout.NewSpacer(),
		createHeader("Configuration"),
		settingsRow,
		layout.NewSpacer(),
		createHeader("Options"),
		optionsRow,
		layout.NewSpacer(),
		createHeader("Progress"),
		progressSection,
	)

	return container.NewVScroll(content)
}

func (g *GuiApp) createLogsTab() fyne.CanvasObject {
	g.logOutput = widget.NewMultiLineEntry()
	g.logOutput.SetPlaceHolder("Logs will appear here...")
	g.logOutput.Disable()

	clearButton := widget.NewButton("Clear", func() {
		g.logOutput.SetText("")
	})
	clearButton.Importance = widget.LowImportance

	header := widget.NewLabelWithStyle("Download Log", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	return container.NewBorder(
		container.NewHBox(
			header,
			layout.NewSpacer(),
			clearButton,
		),
		nil, nil, nil,
		container.NewScroll(g.logOutput),
	)
}

func (g *GuiApp) log(message string) {
	timestamp := time.Now().Format("15:04:05")
	logLine := fmt.Sprintf("[%s] %s\n", timestamp, message)
	g.logOutput.SetText(g.logOutput.Text + logLine)

	// Auto-scroll to bottom
	g.logOutput.CursorRow = len(g.logOutput.Text)
}

func (g *GuiApp) startProcessing() {
	if g.isProcessing {
		return
	}

	// Validate input
	if g.inputFile.Text == "" {
		dialog.ShowError(fmt.Errorf("please select an input file"), g.window)
		return
	}

	if _, err := os.Stat(g.inputFile.Text); os.IsNotExist(err) {
		dialog.ShowError(fmt.Errorf("input file does not exist"), g.window)
		return
	}

	g.isProcessing = true
	g.startButton.Disable()
	g.progressBar.SetValue(0)
	g.statusLabel.SetText("Starting...")
	g.tabs.SelectIndex(1) // Switch to logs tab

	go g.processMemories()
}

func (g *GuiApp) processMemories() {
	defer func() {
		g.isProcessing = false
		g.startButton.Enable()
	}()

	// Parse workers count
	workers := runtime.NumCPU()
	fmt.Sscanf(g.workers.Text, "%d", &workers)
	if workers < 1 {
		workers = 1
	}

	cfg := app.Config{
		InputFile:        g.inputFile.Text,
		OutputDir:        g.outputDir.Text,
		Concurrency:      workers,
		SkipImageOverlay: g.skipImageCheck.Checked,
		SkipVideoOverlay: g.skipVideoCheck.Checked,
		KeepArchives:     g.keepArchCheck.Checked,
		DateFormat:       g.dateFormat.Text,
	}

	g.log(fmt.Sprintf("Starting download with %d workers", workers))
	g.log(fmt.Sprintf("Input file: %s", cfg.InputFile))
	g.log(fmt.Sprintf("Output directory: %s", cfg.OutputDir))

	// Read and parse input file
	content, err := os.ReadFile(cfg.InputFile)
	if err != nil {
		g.log(fmt.Sprintf("ERROR: Failed to read input file: %v", err))
		dialog.ShowError(err, g.window)
		return
	}

	var memories []app.MemoryItem
	ext := filepath.Ext(cfg.InputFile)

	g.log(fmt.Sprintf("Parsing %s file...", ext))

	switch ext {
	case ".html":
		memories = app.ParseHTML(string(content))
	case ".json":
		memories, err = app.ParseJSON(content)
		if err != nil {
			g.log(fmt.Sprintf("ERROR: Failed to parse JSON: %v", err))
			dialog.ShowError(err, g.window)
			return
		}
	default:
		err := fmt.Errorf("unsupported file type: %s", ext)
		g.log(fmt.Sprintf("ERROR: %v", err))
		dialog.ShowError(err, g.window)
		return
	}

	total := len(memories)
	if total == 0 {
		g.log("No memories found in file")
		g.statusLabel.SetText("No memories found")
		dialog.ShowInformation("Complete", "No memories found in the file", g.window)
		return
	}

	g.log(fmt.Sprintf("Found %d memories to download", total))
	g.statusLabel.SetText(fmt.Sprintf("Processing 0/%d", total))

	// Setup worker pool
	var wg sync.WaitGroup
	jobs := make(chan app.MemoryItem, total)
	progressChan := make(chan int, total)

	for w := 1; w <= cfg.Concurrency; w++ {
		wg.Add(1)
		go g.worker(jobs, progressChan, &wg, cfg)
	}

	for _, m := range memories {
		jobs <- m
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(progressChan)
	}()

	// Monitor progress
	completed := 0
	startTime := time.Now()

	for range progressChan {
		completed++
		progress := float64(completed) / float64(total)
		g.progressBar.SetValue(progress)

		elapsed := time.Since(startTime)
		itemsPerSecond := float64(completed) / elapsed.Seconds()
		remainingItems := float64(total - completed)
		var eta time.Duration
		if itemsPerSecond > 0 {
			eta = time.Duration(remainingItems/itemsPerSecond) * time.Second
		}

		statusText := fmt.Sprintf("Processing %d/%d (ETA: %s)", completed, total, eta.Round(time.Second))
		g.statusLabel.SetText(statusText)

		if g.debugCheck.Checked && completed%10 == 0 {
			g.log(fmt.Sprintf("Progress: %d/%d completed", completed, total))
		}
	}

	g.log(fmt.Sprintf("Download complete! Processed %d memories in %s", total, time.Since(startTime).Round(time.Second)))
	g.statusLabel.SetText(fmt.Sprintf("Complete: %d/%d", total, total))
	g.progressBar.SetValue(1.0)

	dialog.ShowInformation("Complete", fmt.Sprintf("Successfully downloaded %d memories!", total), g.window)
}

func (g *GuiApp) worker(jobs <-chan app.MemoryItem, progress chan<- int, wg *sync.WaitGroup, cfg app.Config) {
	defer wg.Done()
	for item := range jobs {
		if g.debugCheck.Checked {
			g.log(fmt.Sprintf("Processing: %s %s", item.Type, item.Date.Format("2006-01-02")))
		}
		app.ProcessItem(item, cfg)
		progress <- 1
	}
}

// Modern theme with optimized font sizes
type modernTheme struct{}

func (m *modernTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(name, variant)
}

func (m *modernTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *modernTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m *modernTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameText:
		return 12
	default:
		return theme.DefaultTheme().Size(name)
	}
}
