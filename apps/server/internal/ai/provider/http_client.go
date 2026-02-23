package provider

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

func newHTTPClient(timeout time.Duration, proxy string) *http.Client {
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	proxy = strings.TrimSpace(proxy)
	if proxy != "" {
		if !strings.Contains(proxy, "://") {
			proxy = "http://" + proxy
		}
		if proxyURL, err := url.Parse(proxy); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}
	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}
