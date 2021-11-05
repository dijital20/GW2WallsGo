package gw2walls

import (
	"sync"
	"time"

	"github.com/juju/loggo"
)

var dl_log = loggo.GetLogger("gw2walls.downloading")

type semaphore chan string

// Downloader is a variant of a sync.WaitGroup.
type Downloader struct {
	sync.WaitGroup
	sem semaphore
}

// processLinks will process the WallpaperLink items from the channel links that match the dimensions string.
func (p *Downloader) processLinks(links chan WallpaperLink, path string, dimensions string) {
	p.Add(1)
	defer p.Done()

	for {
		link, status := <-links
		if !status {
			break
		}

		if link.Dimension == dimensions {
			go p.downloadWallpaper(link, path)

		}
	}

	dl_log.Debugf("Queue closed.")
}

// downloadWallpaper will download a given WallpaperLink to the specified path.
// This is in a separate method so it could be go called.
func (p *Downloader) downloadWallpaper(link WallpaperLink, path string) {
	p.Add(1)
	defer p.Done()

	p.sem <- link.URL

	dl_log.Debugf("Downloading: %s (sem %d)", link.URL, len(p.sem))
	startedAt := time.Now()
	dlPath, err := link.Download(path)
	elapsed := time.Since(startedAt)

	if err != nil {
		dl_log.Errorf("ERROR downloading %s (%s)\n%s\n", link, elapsed, err)
	} else {
		dl_log.Infof("Downloaded %s (%s)", dlPath, elapsed)
	}

	<-p.sem
	dl_log.Debugf("Finished (sem %d)", len(p.sem))
}

// DownloadWallpapers starts a new Downloader instance with a links channel, outputPath and dimensions string. The
// returned Downloader instance can then be Wait() called to wait for it to complete.
func DownloadWallpapers(links chan WallpaperLink, outputPath, dimensions string, cores int) *Downloader {
	var d Downloader

	dl_log.Debugf("Setting up a downloader semaphore with a max of %d", cores)
	d.sem = make(semaphore, cores)

	go d.processLinks(links, outputPath, dimensions)

	return &d
}
