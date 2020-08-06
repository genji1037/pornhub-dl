package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
)

// Video stores informations about a single video on the platform.
type Video struct {
	url, title string
	qualities  map[string]VideoQuality
}

// VideoQuality stores informations about a single file of a specific video.
type VideoQuality struct {
	quality, url, filename string
	filesize               uint64
	ranges                 bool
}

// GetVideoDetails queries the given URL and returns details such as
// - title
// - available qualities
func GetVideoDetails(url string, cfg *Config) (Video, error) {
	// Prepare regex-rules
	slashRegexRule := "\\\\/"
	videoRegexRule := "https://[a-zA-Z]{2}.phncdn.com/videos/[0-9]{6}/[0-9]{2}/[0-9]+/(([0-9]{3,4})P_[0-9]{3,4}K_[0-9]+.mp4)\\?[a-zA-Z0-9%=&_-]{0,210}"
	titleRegexRule := "<title>(.*) - Pornhub.com</title>"

	// Compile regex rules
	slashRegex, _ := regexp.Compile(slashRegexRule)
	videoRegex, _ := regexp.Compile(videoRegexRule)
	titleRegex, _ := regexp.Compile(titleRegexRule)

	// Download content of webpage
	// resp, err := http.Get(url)
	resp, err := getResp(url, cfg)

	// Check if there was an error downloading
	if err != nil {
		return Video{}, err
	}

	// Get body and manipulate it to make searching with regex easiert
	defer resp.Body.Close()
	source, err := ioutil.ReadAll(resp.Body)
	body := slashRegex.ReplaceAllString(string(source), "/")

	// Get title
	title := titleRegex.FindStringSubmatch(body)[1]

	// Get all videos from the website
	videoUrls := videoRegex.FindAllStringSubmatch(body, -1)
	videoQualities := make(map[string]VideoQuality)

	// Iterate through all found URLs to process them
	for _, slice := range videoUrls {
		// Continue if the result does not contain 3 items (url, filename, quality)
		if len(slice) != 3 {
			continue
		}

		// Get data
		url := slice[0]
		filename := slice[1]
		quality := slice[2]
		ranges := false
		var filesize uint64

		// Do Head-Request to the file to fetch some data
		headResp, err := http.Head(url)
		if err != nil {
			fmt.Println("Error while preparing download:")
			fmt.Println(err)
		}
		defer headResp.Body.Close()

		// Get size of file
		filesize, _ = strconv.ParseUint(headResp.Header.Get("Content-Length"), 10, 64)

		// Check if the server supports the Range-header
		if headResp.Header.Get("Accept-Ranges") == "bytes" {
			ranges = true
		}

		// Create and store video quality object
		videoQuality := VideoQuality{url: url, filename: filename, quality: quality, filesize: filesize, ranges: ranges}
		videoQualities[quality] = videoQuality
	}

	// Check if there are videos on the website. If not, cancel
	if len(videoQualities) == 0 {
		f, _ := os.Create("site.html")
		f.WriteString(body)
		return Video{}, errors.New("could not find any video sources")
	}

	// Sort qualities descending
	var qualityKeys []string
	for k := range videoQualities {
		qualityKeys = append(qualityKeys, k)
	}
	sort.Strings(qualityKeys)

	sortedQualities := make(map[string]VideoQuality)
	for _, k := range qualityKeys {
		sortedQualities[k] = videoQualities[k]
	}

	// Return the video detail instance
	video := Video{url: url, title: title, qualities: videoQualities}
	return video, nil
}
