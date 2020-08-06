package main

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"sync/atomic"
)

// DownloadStatus counts the written bytes. Because it implements the io.Writer interface,
// it can be given to the io.TeeReader(). This is also used to print out the current
// status of the download
type DownloadStatus struct {
	DoneBytes   uint64
	TotalBytes  uint64
	DoneThread  uint32
	TotalThread uint32
}

// Write implements io.Write
func (s *DownloadStatus) Write(bytes []byte) (int, error) {
	// Count the amount of bytes written since the last cycle
	byteAmount := len(bytes)

	// Increment current count by the amount of bytes written since the last cycle
	atomic.AddUint64(&s.DoneBytes, uint64(byteAmount))

	// Update progress
	s.Print()

	// Return byteAmount
	return byteAmount, nil
}

// Print prints the current download progress to console.
func (s DownloadStatus) Print() {
	fmt.Printf("\rDownloading %s / %s (%d of %d threads done) ",
		humanize.Bytes(s.DoneBytes), humanize.Bytes(s.TotalBytes), s.DoneThread, s.TotalThread)
}
