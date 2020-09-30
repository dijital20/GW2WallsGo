// Program for finding and downloading GuildWars 2 Wallpapers.

package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly"
	"github.com/juju/loggo"
	"github.com/juju/loggo/loggocolor"
	"github.com/pbnjay/strptime"
)

const (
	mediaURL    = "https://www.guildwars2.com/en/media/wallpapers/"
	releasesURL = "https://www.guildwars2.com/en/the-game/releases/"
	mainSiteURL = "https://www.guildwars2.com"
)

var (
	seeders sync.WaitGroup
	pullers sync.WaitGroup
	log     = loggo.GetLogger("GW2Wall")
)

type wallpaperLink struct {
	URL       string
	Release   string
	Dimension string
	Date      string
	Number    int
}

// Download the wallpaper to the local machine.
func (l *wallpaperLink) Download(path string) (string, error) {
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

	// Make the output dir
	os.MkdirAll(filepath.Dir(dst), 0666)

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
func (l *wallpaperLink) String() string {
	fixer, _ := regexp.Compile("[^\\w\\s\\-]")
	var name string
	if l.Date == "" {
		name = fmt.Sprintf("%s %d %s", l.Release, l.Number, l.Dimension)
	} else {
		name = fmt.Sprintf("%s %s %d %s", l.Date, l.Release, l.Number, l.Dimension)
	}
	return fixer.ReplaceAllString(name, "")
}

// Parse the releases page to get the links to each release page.
// Calls getLinks for each release page found.
func getReleases(url string, links chan wallpaperLink) {
	log.Infof("Getting releases from: %s", url)

	defer seeders.Done()
	c := colly.NewCollector()

	c.OnHTML("section.release-canvas", func(release *colly.HTMLElement) {
		release.ForEach("li", func(_ int, item *colly.HTMLElement) {
			item.ForEach("a", func(_ int, sizeItem *colly.HTMLElement) {
				rurl := sizeItem.Attr("href")
				if !strings.HasPrefix(rurl, mainSiteURL) {
					rurl = mainSiteURL + rurl
				}

				go getLinks(rurl, links)
				seeders.Add(1)
			})
		})
	})

	c.Visit(url)
}

// Parses the specified page for its sweet, sweet wallpaper download links.
// Populates a channel with a wallpaperLink object which has the URL,
// release name, release date, and dimensions.
func getLinks(url string, links chan wallpaperLink) {
	log.Infof("Getting links from: %s", url)

	defer seeders.Done()
	c := colly.NewCollector()
	var (
		name        string
		wallNumber  int
		releaseDate string
	)

	// Title
	c.OnHTML("title", func(title *colly.HTMLElement) {

		// Set release name
		name = strings.TrimSpace(strings.Replace(title.Text, " | GuildWars2.com", "", 1))

		// Set release date
		urlSegments := strings.Split(url, "/")
		dateSegment := urlSegments[len(urlSegments)-2]
		for _, dateFormat := range []string{"%B-%Y", "%B-%d-%Y"} {
			parsedDate, err := strptime.Parse(dateSegment, dateFormat)
			if err != nil {
				continue
			}
			releaseDate = fmt.Sprintf("%04d-%02d", parsedDate.Year(), parsedDate.Month())
			break
		}

		log.Debugf("Set name and date:", name, releaseDate)
	})

	// Media URLs
	c.OnHTML("li.wallpaper", func(item *colly.HTMLElement) {
		wallNumber++

		item.ForEach("img", func(_ int, imgItem *colly.HTMLElement) {
			src := strings.Split(item.Attr("src"), "/")
			name = strings.Replace(src[len(src)-1], "-crop.jpg", "", 1)
			log.Debugf("Set name to:", name)
		})

		item.ForEach("a", func(_ int, sizeItem *colly.HTMLElement) {
			link := wallpaperLink{
				URL:       sizeItem.Attr("href"),
				Dimension: strings.TrimSpace(sizeItem.Text),
				Release:   "Media",
				Number:    wallNumber,
			}

			log.Debugf("Found %s", link.String())
			links <- link
		})
	})

	// Release Walls
	c.OnHTML("ul.wallpaper", func(item *colly.HTMLElement) {
		wallNumber++

		item.ForEach("a", func(_ int, sizeItem *colly.HTMLElement) {
			wURL := sizeItem.Attr("href")
			if !strings.HasPrefix(wURL, "https:") {
				wURL = "https:" + wURL
			}
			link := wallpaperLink{
				URL:       wURL,
				Dimension: sizeItem.Text,
				Release:   name,
				Date:      releaseDate,
				Number:    wallNumber,
			}

			log.Debugf("Found %s", link.String())
			links <- link
		})
	})

	c.OnHTML("ul.resolution", func(item *colly.HTMLElement) {
		wallNumber++

		item.ForEach("a", func(_ int, sizeItem *colly.HTMLElement) {
			wURL := sizeItem.Attr("href")
			if !strings.HasPrefix(wURL, "https:") {
				wURL = "https:" + wURL
			}
			link := wallpaperLink{
				URL:       wURL,
				Dimension: sizeItem.Text,
				Release:   name,
				Date:      releaseDate,
				Number:    wallNumber,
			}

			log.Debugf("Found %s", link.String())
			links <- link
		})
	})

	c.Visit(url)
}

// Processes the download links in the channel to save them locally.
func processLinks(links chan wallpaperLink, path string, dimensions string) {
	defer pullers.Done()

	for {
		link, status := <-links
		if !status {
			break
		}

		if link.Dimension == dimensions {
			log.Debugf("Downloading: %s", link.URL)
			startedAt := time.Now()
			dlPath, err := link.Download(path)
			elapsed := time.Now().Sub(startedAt)
			log.Debugf("Downloaded in %s", elapsed)
			if err != nil {
				log.Errorf("ERROR downloading %s\n%s\n", link, err)
			} else {
				log.Infof("Downloaded %s", dlPath)
			}
		}
	}
}

func main() {
	// CLI Parsing.
	dimensions := flag.String("dimension", "1920x1080", "Dimensions of the wallpapers to download.")
	outputPath := flag.String("output-path", "gw2_walls", "Path to download files to.")
	verbose := flag.Bool("verbose", false, "Enable verbose logging.")
	skipMedia := flag.Bool("skip-media", false, "Skip Media wallpapers.")
	skipRelease := flag.Bool("skip-release", false, "Skip Release wallpapers.")
	flag.Parse()
	startedAt := time.Now()

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

	links := make(chan wallpaperLink, 200)

	// Start up the seeders... they find things to download.
	if !*skipRelease {
		go getReleases(releasesURL, links)
		seeders.Add(1)
	}
	if !*skipMedia {
		go getLinks(mediaURL, links)
		seeders.Add(1)
	}

	// Start up the puller... they download things.
	go processLinks(links, *outputPath, *dimensions)
	pullers.Add(1)

	// When the seeders finish, close the channel the pullers are using.
	seeders.Wait()
	close(links)

	// When the pullers are done, then you can exit.
	pullers.Wait()

	elapsed := time.Now().Sub(startedAt)
	log.Infof("Finished in %s seconds.", elapsed)
}
