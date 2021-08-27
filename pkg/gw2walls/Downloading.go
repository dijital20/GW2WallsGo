package gw2walls

import (
	"sync"
	"time"

	"github.com/juju/loggo"
)

var dl_log = loggo.GetLogger("gw2walls.downloading")

type Downloader struct {
	sync.WaitGroup
}

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
}

func (p *Downloader) downloadWallpaper(link WallpaperLink, path string) {
	p.Add(1)
	defer p.Done()

	dl_log.Debugf("Downloading: %s", link.URL)
	startedAt := time.Now()
	dlPath, err := link.Download(path)
	elapsed := time.Now().Sub(startedAt)
	dl_log.Debugf("Downloaded in %s", elapsed)
	if err != nil {
		dl_log.Errorf("ERROR downloading %s\n%s\n", link, err)
	} else {
		dl_log.Infof("Downloaded %s", dlPath)
	}
}

func DownloadWallpapers(links chan WallpaperLink, outputPath, dimensions string) *Downloader {
	var d Downloader
	go d.processLinks(links, outputPath, dimensions)

	return &d
}
