package gw2walls

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/juju/loggo"
)

var wl_log = loggo.GetLogger("gw2walls.wallpaper")

// WallpaperLink represents a found wallpaper link. It contains the URL to the file, the Release it came from,
// Dimensions of it, Date it was released, and index Number.
type WallpaperLink struct {
	URL       string
	Release   string
	Dimension string
	Date      string
	Number    int
}

// Download the wallpaper to the local machine.
func (l *WallpaperLink) Download(path string) (string, error) {
	ext := filepath.Ext(l.URL)
	if ext == "" {
		ext = ".jpg"
	}
	dst := filepath.Join(path, l.String()+ext)

	// Open the file download.
	resp, err := http.Get(l.URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check if file exists, and log warning if it does.
	if _, err := os.Stat(dst); err == nil {
		wl_log.Warningf("Path %s already exists, overwriting.", dst)
	}

	// Make the output dir
	os.MkdirAll(filepath.Dir(dst), 0777)

	// Open the local file.
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Copy the file contents from file to file.
	_, err = io.Copy(out, resp.Body)
	return dst, err
}

// Get a nice string representation of the release name and date that is filename friendly.
func (l *WallpaperLink) String() string {
	fixer, _ := regexp.Compile(`[^\w\s\-]`)
	var name string
	if l.Date == "" {
		name = fmt.Sprintf("%s %d %s", l.Release, l.Number, l.Dimension)
	} else {
		name = fmt.Sprintf("%s %s %d %s", l.Date, l.Release, l.Number, l.Dimension)
	}
	return fixer.ReplaceAllString(name, "")
}
