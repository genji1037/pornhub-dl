package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/dustin/go-humanize"
)

func main() {

	cfg := NewConfig()

	// Retrieve video details
	videoDetails, err := GetVideoDetails(cfg.URL, cfg)
	if err != nil {
		log.Fatalf("GetVideoDetails failed: %v", err)
	}

	// Process given quality
	var highestQuality int64
	selectedQualityName := cfg.Quality
	if cfg.Quality == "highest" {
		for i := range videoDetails.qualities {
			currentQuality, _ := strconv.ParseInt(i, 10, 64)
			if currentQuality > highestQuality {
				highestQuality = currentQuality
			}
		}

		selectedQualityName = strconv.FormatInt(highestQuality, 10)
	}

	// Print video details
	fmt.Println("Title: " + videoDetails.title + "\n")
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 6, 8, 2, ' ', 0)

	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t\n", "Quality", "Filename", "Size", "FastDL", "")
	//fmt.Fprintf(w, "\n%s\t%s\t%s\t%s\t%s\t", "-------", "--------", "----", "------", "")

	for _, quality := range videoDetails.qualities {
		x := " "
		if selectedQualityName == quality.quality {
			x = "←"
		}

		fmt.Fprintf(w, "%sp\t%s\t%s\t%t\t%s\t\n", quality.quality, quality.filename, humanize.Bytes(quality.filesize), quality.ranges, x)
	}

	w.Flush()

	if _, ok := videoDetails.qualities[selectedQualityName]; !ok {
		fmt.Println("Quality " + selectedQualityName + " is not available for this video.")
		return
	}

	selectedQuality := videoDetails.qualities[selectedQualityName]

	fmt.Println("")

	if cfg.OutputPath == "default" {
		cfg.OutputPath = selectedQuality.filename
	}

	if selectedQuality.ranges {
		SplitDownloadFile(cfg.OutputPath, selectedQuality, cfg)
	} else {
		DownloadFile(cfg.OutputPath, selectedQuality.url, cfg)
	}
}
