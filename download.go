package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
)

type PartialDownloadSource struct {
	URL    string
	Offset uint64
	End    uint64
}

// DoPartialDownload downloads a special part of the file at the given URL.
func DoPartialDownload(src PartialDownloadSource, output *os.File, counter *DownloadStatus) error {
	client := http.Client{}

	// Build request
	req, _ := http.NewRequest("GET", src.URL, nil)
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", src.Offset, src.End))
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Referer", src.URL)

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create our progress reporter and pass it to be used alongside our writer
	_, err = io.Copy(output, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	output.Close()
	return nil
}

// SplitDownloadFile downloads a remote file to the harddrive while writing it
// directly to a file instead of storing it in RAM until the donwload completes.
func SplitDownloadFile(filepath string, video VideoQuality, cfg *Config) error {
	downloadStatus := &DownloadStatus{
		TotalBytes:  video.filesize,
		TotalThread: uint32(cfg.ThreadNum),
	}
	sliceSize := video.filesize / uint64(cfg.ThreadNum)

	allWorkOK := true
	wg := sync.WaitGroup{}
	wg.Add(cfg.ThreadNum)
	for i := 1; i <= cfg.ThreadNum; i++ {
		offset := sliceSize * uint64(i-1)
		end := offset + sliceSize - 1

		if i == cfg.ThreadNum {
			end = video.filesize
		}

		// Create a temporary file
		tempfileName := fmt.Sprintf("%s.%d.tmp", filepath, i)
		output, err := os.Create(tempfileName)
		if err != nil {
			return err
		}

		src := PartialDownloadSource{
			URL:    video.url,
			Offset: offset,
			End:    end,
		}

		go func() {
			err = RetryOnError(func() error {
				return DoPartialDownload(src, output, downloadStatus)
			}, 3, "DoPartialDownload")
			if err != nil {
				allWorkOK = false
			}
			atomic.AddUint32(&downloadStatus.DoneThread, 1)
			wg.Done()
		}()
	}

	fmt.Printf("Downloading file using %d threads.\n", cfg.ThreadNum)
	wg.Wait()
	downloadStatus.Print()
	fmt.Print("combining temporary file... ")

	if !allWorkOK {
		// TODO persist failed part in order to retry download.
		return fmt.Errorf("failed to download some part")
	}
	// Combine single downloads into a single video file
	output, _ := os.Create(filepath)
	defer output.Close()
	for i := 1; i <= cfg.ThreadNum; i++ {
		// Open temporary file
		tempfileName := fmt.Sprintf("%s.%d.tmp", filepath, i)
		file, _ := os.Open(tempfileName)

		// Read bytes
		stat, _ := file.Stat()
		tempBytes := make([]byte, stat.Size())
		file.Read(tempBytes)

		// Write to output file
		output.Write(tempBytes)

		// Close and delete temporary file
		file.Close()

		err := os.Remove(tempfileName)
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("DoneBytes!")
	return nil
}

// DownloadFile downloads a remote file to the harddrive while writing it
// directly to a file instead of storing it in RAM until the donwload completes.
func DownloadFile(filepath string, url string, cfg *Config) error {
	fmt.Println("Server does not support partial downloads. Continuing with a single thread.")

	// Create a temporary file
	tempfile := filepath + ".tmp"
	output, err := os.Create(tempfile)
	if err != nil {
		return err
	}
	defer output.Close()

	// Download data from given URL
	// resp, err := http.Get(url)
	resp, err := getResp(url, cfg)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read total size of file
	filesize, _ := strconv.ParseUint(resp.Header.Get("Content-Length"), 10, 64)

	// Create our progress reporter and pass it to be used alongside our writer
	counter := &DownloadStatus{TotalBytes: filesize}
	_, err = io.Copy(output, io.TeeReader(resp.Body, counter))
	if err != nil {
		return err
	}

	// Rename temp file to correct ending
	err = os.Rename(tempfile, filepath)
	if err != nil {
		return err
	}

	return nil
}
