# GW2WallsGo
Guild Wars 2 Wallpaper downloader, written in Go. This program is an improvement on my previous Python-based solution, which now includes concurrency.

The program works by scraping links from the various pages with wallpapers, and then downloads the ones that match the specified dimensions.

## How to build

1. Install `go`
2. Download or clone this repository.
3. In the directory you have cloned or unzipped this code to, run `go build GW2WallsGo.go`.

## How to use

`GW2WallsGo` has these options:

```
Usage of C:\Users\joshs\Projects\GW2WallsGo\GW2WallsGo.exe:
  -dimension string
        Dimensions of the wallpapers to download. (default "1920x1080")
  -output-path string
        Path to download files to. (default "gw2_walls")
  -skip-media
        Skip Media wallpapers.
  -skip-release
        Skip Release wallpapers.
  -verbose
        Enable verbose logging.
```

Runs pretty fast. I did this script just to learn Go.
