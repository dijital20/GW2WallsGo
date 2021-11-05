package gw2walls

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/juju/loggo"
	"github.com/pbnjay/strptime"
)

const (
	mediaURL    = "https://www.guildwars2.com/en/media/wallpapers/"
	releasesURL = "https://www.guildwars2.com/en/the-game/releases/"
	mainSiteURL = "https://www.guildwars2.com"
)

var sc_log = loggo.GetLogger("gw2walls.scrape")

// Scraper type contains a WaitGroup. Can call Wait to see when it is done.
type Scraper struct {
	sync.WaitGroup
}

// Parse the releases page to get the links to each release page.
// Calls getLinks for each release page found.
func (s *Scraper) getReleases(url string, links chan WallpaperLink) {
	s.Add(1)
	defer s.Done()

	sc_log.Infof("Getting releases from: %s", url)
	c := colly.NewCollector()

	c.OnHTML("section.release-canvas", func(release *colly.HTMLElement) {
		release.ForEach("li", func(_ int, item *colly.HTMLElement) {
			item.ForEach("a", func(_ int, sizeItem *colly.HTMLElement) {
				rurl := sizeItem.Attr("href")
				if !strings.HasPrefix(rurl, mainSiteURL) {
					rurl = mainSiteURL + rurl
				}

				go s.getLinks(rurl, links)
			})
		})
	})

	c.Visit(url)
	sc_log.Debugf("Finished processing %s", url)
}

// Parses the specified page for its sweet, sweet wallpaper download links.
// Populates a channel with a WallpaperLink object which has the URL,
// release name, release date, and dimensions.
func (s *Scraper) getLinks(url string, links chan WallpaperLink) {
	s.Add(1)
	defer s.Done()

	sc_log.Infof("Getting links from: %s", url)
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

		sc_log.Debugf("Set name and date:", name, releaseDate)
	})

	// Media URLs
	c.OnHTML("li.wallpaper", func(item *colly.HTMLElement) {
		wallNumber++

		item.ForEach("img", func(_ int, imgItem *colly.HTMLElement) {
			src := strings.Split(item.Attr("src"), "/")
			name = strings.Replace(src[len(src)-1], "-crop.jpg", "", 1)
			sc_log.Debugf("Set name to:", name)
		})

		item.ForEach("a", func(_ int, sizeItem *colly.HTMLElement) {
			link := WallpaperLink{
				URL:       sizeItem.Attr("href"),
				Dimension: strings.TrimSpace(sizeItem.Text),
				Release:   "Media",
				Number:    wallNumber,
			}

			sc_log.Debugf("Found %s", link.String())
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
			link := WallpaperLink{
				URL:       wURL,
				Dimension: sizeItem.Text,
				Release:   name,
				Date:      releaseDate,
				Number:    wallNumber,
			}

			sc_log.Debugf("Found %s", link.String())
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
			link := WallpaperLink{
				URL:       wURL,
				Dimension: sizeItem.Text,
				Release:   name,
				Date:      releaseDate,
				Number:    wallNumber,
			}

			sc_log.Debugf("Found %s", link.String())
			links <- link
		})
	})

	c.Visit(url)
	sc_log.Debugf("Finished processing %s", url)
}

// Starts the process of finding wallpapers. If skipReleases, skips the "release" wallpapers, and if shipMedia, skips
// the "media" wallpapers. Returns a pointer to a channel that will be filled with found WallpaperLink objects, and a
// pointer to the scraper instance, which we can use to tell when the scraping is done.
func FindWallpapers(skipReleases, skipMedia bool) (*chan WallpaperLink, *Scraper) {
	var s Scraper
	links := make(chan WallpaperLink, 200)

	if !skipReleases {
		sc_log.Debugf("Calling getReleases")
		go s.getReleases(releasesURL, links)
	}

	if !skipMedia {
		sc_log.Debugf("Starting getLinks")
		go s.getLinks(mediaURL, links)
	}

	return &links, &s
}
