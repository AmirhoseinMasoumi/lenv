package progress

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

// Bar represents a progress bar for tracking download/operation progress.
type Bar struct {
	total       int64
	current     int64
	startTime   time.Time
	lastUpdate  time.Time
	description string
	width       int
	mu          sync.Mutex
	done        bool
	isTTY       bool
}

// NewBar creates a new progress bar with the given total size and description.
func NewBar(total int64, description string) *Bar {
	width := 40
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 60 {
		width = min(w-40, 60)
	}

	isTTY := term.IsTerminal(int(os.Stdout.Fd()))

	return &Bar{
		total:       total,
		description: description,
		startTime:   time.Now(),
		lastUpdate:  time.Now(),
		width:       width,
		isTTY:       isTTY,
	}
}

// Writer wraps an io.Writer and tracks progress.
type Writer struct {
	io.Writer
	bar *Bar
}

// NewWriter creates a progress-tracking writer wrapper.
func NewWriter(w io.Writer, bar *Bar) *Writer {
	return &Writer{Writer: w, bar: bar}
}

func (pw *Writer) Write(p []byte) (int, error) {
	n, err := pw.Writer.Write(p)
	if n > 0 {
		pw.bar.Add(int64(n))
	}
	return n, err
}

// Reader wraps an io.Reader and tracks progress.
type Reader struct {
	io.Reader
	bar *Bar
}

// NewReader creates a progress-tracking reader wrapper.
func NewReader(r io.Reader, bar *Bar) *Reader {
	return &Reader{Reader: r, bar: bar}
}

func (pr *Reader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	if n > 0 {
		pr.bar.Add(int64(n))
	}
	return n, err
}

// Add increments the current progress by n bytes.
func (b *Bar) Add(n int64) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.current += n
	now := time.Now()

	// Throttle updates to every 100ms for TTY, or every 5% for non-TTY
	if b.isTTY {
		if now.Sub(b.lastUpdate) < 100*time.Millisecond && b.current < b.total {
			return
		}
	} else {
		percent := float64(b.current) / float64(b.total) * 100
		lastPercent := float64(b.current-n) / float64(b.total) * 100
		if int(percent/5) == int(lastPercent/5) && b.current < b.total {
			return
		}
	}

	b.lastUpdate = now
	b.render()
}

// Finish completes the progress bar.
func (b *Bar) Finish() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.done {
		return
	}
	b.done = true
	b.current = b.total
	b.render()
	fmt.Println()
}

func (b *Bar) render() {
	if b.total <= 0 {
		return
	}

	percent := float64(b.current) / float64(b.total) * 100
	elapsed := time.Since(b.startTime).Seconds()

	// Calculate speed
	var speed float64
	if elapsed > 0 {
		speed = float64(b.current) / elapsed
	}

	// Calculate ETA
	var eta string
	if speed > 0 && b.current < b.total {
		remaining := float64(b.total-b.current) / speed
		eta = formatDuration(time.Duration(remaining) * time.Second)
	} else if b.current >= b.total {
		eta = "done"
	} else {
		eta = "calculating..."
	}

	if b.isTTY {
		b.renderTTY(percent, speed, eta)
	} else {
		b.renderNonTTY(percent, speed)
	}
}

func (b *Bar) renderTTY(percent, speed float64, eta string) {
	// Build progress bar
	filled := int(percent / 100 * float64(b.width))
	if filled > b.width {
		filled = b.width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", b.width-filled)

	// Format sizes
	currentStr := formatBytes(b.current)
	totalStr := formatBytes(b.total)
	speedStr := formatBytes(int64(speed)) + "/s"

	// Clear line and print
	fmt.Printf("\r\x1b[K%s [%s] %5.1f%% %s/%s %s ETA: %s",
		b.description, bar, percent, currentStr, totalStr, speedStr, eta)
}

func (b *Bar) renderNonTTY(percent, speed float64) {
	currentStr := formatBytes(b.current)
	totalStr := formatBytes(b.total)
	speedStr := formatBytes(int64(speed)) + "/s"

	fmt.Printf("  %s: %.0f%% (%s/%s) %s\n",
		b.description, percent, currentStr, totalStr, speedStr)
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		m := int(d.Minutes())
		s := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", m, s)
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", h, m)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
