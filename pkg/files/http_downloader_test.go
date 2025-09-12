package files

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pixie-sh/core-go/pkg/base64"
)

func TestHTTPDownloader_Download(t *testing.T) {
	// Setup test server
	expectedContent := "test file content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedContent))
	}))
	defer server.Close()

	// Create downloader with default client
	downloader := NewHTTPDownloader(nil, http.DefaultClient, "CustomUserAgent")

	// Test Download method
	content, err := downloader.Download(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	if !reflect.DeepEqual(content, []byte(expectedContent)) {
		t.Errorf("Expected content %s, got %s", expectedContent, string(content))
	}

	// Test error case with non-existent URL
	_, err = downloader.Download(context.Background(), "http://non-existent-url.example")
	if err == nil {
		t.Error("Expected error for non-existent URL, got nil")
	}
}

func TestHTTPDownloader_Stream(t *testing.T) {
	// Setup test server
	expectedContent := "test file content for streaming"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedContent))
	}))
	defer server.Close()

	// Create downloader with default client
	downloader := NewHTTPDownloader(nil, http.DefaultClient, "CustomUserAgent")

	// Test Stream method
	reader, err := downloader.Stream(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read stream: %v", err)
	}

	if !reflect.DeepEqual(content, []byte(expectedContent)) {
		t.Errorf("Expected content %s, got %s", expectedContent, string(content))
	}

	// Test HTTP error status code
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer errorServer.Close()

	_, err = downloader.Stream(context.Background(), errorServer.URL)
	if err == nil {
		t.Error("Expected error for 404 status, got nil")
	}
}

// Test with a custom client
func TestHTTPDownloaderWithCustomClient(t *testing.T) {
	expectedContent := "custom client test"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the custom headers were passed
		if r.Header.Get("User-Agent") != "CustomUserAgent" {
			t.Errorf("Expected User-Agent header, got none")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedContent))
	}))
	defer server.Close()

	// Create a custom client
	customClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
		},
	}

	// Create downloader with custom client
	downloader := NewHTTPDownloader(nil, customClient, "CustomUserAgent")

	// We need to wrap the Download call to set headers
	ctx := context.Background()
	// This is just to test the custom client works, actual implementation would need modification
	// to support custom headers
	content, err := downloader.Download(ctx, server.URL)

	if err != nil {
		t.Fatalf("Download with custom client failed: %v", err)
	}

	if !strings.Contains(string(content), expectedContent) {
		t.Errorf("Expected content to contain %s, got %s", expectedContent, string(content))
	}
}

// Test with a custom client
func TestHTTPDownloaderRealAsset(t *testing.T) {
	// Create a custom client
	customClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
		},
	}

	downloader := NewHTTPDownloader(nil, customClient, "Pixie/test (go1.24; +https://pixie.sh)")
	ctx := context.Background()
	content, err := downloader.Download(ctx, "https://pixie.sh/assets/Images/pixie-logo.png")

	if err != nil {
		t.Fatalf("Download with custom client failed: %v", err)
	}

	fmt.Printf("b64 content: %s\n", base64.MustEncode(content))
}
