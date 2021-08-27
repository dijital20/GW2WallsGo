package main

import (
	"flag"
	"os"
	"path/filepath"
	"time"

	"github.com/dijital20/GW2WallsGo/pkg/gw2walls"
	"github.com/juju/loggo"
	"github.com/juju/loggo/loggocolor"
)

func main() {
	// CLI Parsing.
	dimensions := flag.String("dimension", "1920x1080", "Dimensions of the wallpapers to download.")
	outputPath := flag.String("output-path", "gw2_walls", "Path to download files to.")
	verbose := flag.Bool("verbose", false, "Enable verbose logging.")
	skipMedia := flag.Bool("skip-media", false, "Skip Media wallpapers.")
	skipRelease := flag.Bool("skip-release", false, "Skip Release wallpapers.")
	flag.Parse()
	startedAt := time.Now()

	log := loggo.GetLogger("gw2walls")

	*outputPath, _ = filepath.Abs(*outputPath)
	if *verbose {
		log.SetLogLevel(loggo.DEBUG)
		loggo.ReplaceDefaultWriter(loggocolor.NewWriter(os.Stdout))
	} else {
		log.SetLogLevel(loggo.INFO)
		loggo.ReplaceDefaultWriter(loggo.NewSimpleWriter(os.Stdout, func(entry loggo.Entry) string {
			return entry.Message
		}))
	}

	log.Infof("Dimensions: %s\nPath: %s\n", *dimensions, *outputPath)

	links, scraper := gw2walls.FindWallpapers(*skipRelease, *skipMedia)
	downloader := gw2walls.DownloadWallpapers(*links, *outputPath, *dimensions)

	log.Debugf("Waiting for scraper to finish...")
	scraper.Wait()

	log.Debugf("Waiting for downloads to complete...")
	downloader.Wait()

	elapsed := time.Now().Sub(startedAt)
	log.Infof("Finished in %s seconds.", elapsed)
}
