package main

import (
	"context"
	"fmt"
	goproxy "golang.org/x/net/proxy"
	"net"
	"net/http"
	"os"
)

const userAgent = `Mozilla/5.0 (Windows NT 6.3; Trident/7.0; rv:11.0) like Gecko`

// Start HTTP GET Request
// - url
// - proxyAddr
func getResp(url string, cfg *Config) (*http.Response, error) {
	httpTransport := &http.Transport{}
	client := &http.Client{Transport: httpTransport}

	if len(cfg.Socks5) > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "Socks5 proxy address is %s\n", cfg.Socks5)
		dialer, err := goproxy.SOCKS5("tcp", cfg.Socks5, nil, goproxy.Direct)
		if err != nil {
			return nil, err
		}

		httpTransport.DialContext = func(ctx context.Context, network, addr string) (conn net.Conn, e error) {
			return dialer.Dial(network, addr)
		}
	}

	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add("User-Agent", userAgent)
	request.Header.Add("Referer", url)
	if err != nil {
		return nil, err
	}

	return client.Do(request)
}
