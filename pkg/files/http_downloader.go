package files

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type HTTPDownloader interface {
	Download(ctx context.Context, filePath string) ([]byte, error)
	Stream(ctx context.Context, filePath string) (io.ReadCloser, error)
}

// httpDownloader implements HTTPDownloader for downloading files from HTTP URLs
type httpDownloader struct {
	client   *http.Client
	uaHeader http.Header
}

// NewHTTPDownloader creates a new httpDownloader with optional custom HTTP client
func NewHTTPDownloader(_ context.Context, client *http.Client, uaHeader string) *httpDownloader {
	return &httpDownloader{
		client: client,
		uaHeader: http.Header{
			"User-Agent": []string{uaHeader},
		},
	}
}

// Download retrieves a file from a URL and returns its content as a byte slice
func (d *httpDownloader) Download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header = d.uaHeader
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// Stream retrieves a file from a URL and returns a ReadCloser for streaming
func (d *httpDownloader) Stream(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header = d.uaHeader
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error performing request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
