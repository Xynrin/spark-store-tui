package preview

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Brand renders the configured logo with chafa. The return value is ANSI art
// that can be embedded directly into a Bubble Tea view. No image dependency
// means an empty return value, allowing the UI to stay fully functional.
func Brand() string {
	return RenderFile(brandImagePath())
}

// RenderURL caches a remote image before asking chafa to render it. Downloads
// are bounded to avoid turning untrusted metadata into an unbounded cache.
func RenderURL(url string) string {
	if url == "" {
		return ""
	}
	cacheDirectory, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	name := fmt.Sprintf("%x.img", sha256.Sum256([]byte(url)))
	path := filepath.Join(cacheDirectory, "sparkstore", "images", name)
	if !isFile(path) {
		if err := fetch(url, path); err != nil {
			return ""
		}
	}
	return RenderFile(path)
}

func RenderFile(path string) string {
	if path == "" {
		return ""
	}
	if _, err := exec.LookPath("chafa"); err != nil {
		return ""
	}
	// TUI output is captured rather than written to chafa's controlling terminal.
	// Probing in that situation can block for seconds on some terminals.
	command := exec.Command("chafa", "--probe=off", "--format=symbols", "--size=18x9", "--animate=off", path)
	output, err := command.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func fetch(url, path string) error {
	client := &http.Client{Timeout: 8 * time.Second}
	response, err := client.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("image request returned %s", response.Status)
	}
	if response.ContentLength > 12<<20 {
		return fmt.Errorf("image is larger than 12 MiB")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary := path + ".tmp"
	file, err := os.OpenFile(temporary, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	limited := &io.LimitedReader{R: response.Body, N: 12<<20 + 1}
	_, copyErr := io.Copy(file, limited)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if limited.N == 0 {
		_ = os.Remove(temporary)
		return fmt.Errorf("image is larger than 12 MiB")
	}
	return os.Rename(temporary, path)
}

func brandImagePath() string {
	if configured := os.Getenv("SPARK_STORE_TUI_BRAND_IMAGE"); configured != "" {
		if isFile(configured) {
			return configured
		}
		return ""
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return ""
	}
	candidate := filepath.Join(workingDirectory, "icon.png")
	if isFile(candidate) {
		return candidate
	}
	return ""
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
