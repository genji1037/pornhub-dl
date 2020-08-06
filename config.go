package main

import (
	"errors"
	"flag"
)

type Config struct {
	URL        string
	Quality    string
	OutputPath string
	Debug      bool
	ThreadNum  int
	Socks5     string
	URLFile    string
}

func NewConfig() *Config {
	config := Config{}
	// Define flags and parse them
	flag.StringVar(&config.URL, "u", "", "URL of the video to download")
	flag.StringVar(&config.Quality, "q", "highest", "The quality number (eg. 720) or 'highest'")
	flag.StringVar(&config.OutputPath, "o", "default", "Path to where the download should be saved or 'default' for the original filename")
	flag.BoolVar(&config.Debug, "d", false, "Whether you want to activate debug mode or not")
	flag.IntVar(&config.ThreadNum, "c", 5, "The amount of threads to use to download")
	flag.StringVar(&config.Socks5, "s", "", "Specify socks5 proxy address for downloading resources")
	flag.StringVar(&config.URLFile, "f", "", "Specify file path of URLs, URLs must separate by \\n")
	flag.Parse()
	return &config
}

func (c *Config) Validate() error {
	// Check if parameters are set
	if c.URL == "" && c.URLFile == "" {
		return errors.New("please pass a valid url with the -u parameter or file that contain URLs with the -f parameter")
	}
	if c.URL != "" && c.URLFile != "" {
		return errors.New("both -u -f specified, choose one of them please")
	}
	return nil
}
