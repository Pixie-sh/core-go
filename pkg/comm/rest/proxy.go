package rest

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	goHttp "net/http"

	"github.com/andybalholm/brotli"
	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/env"

	pixieEnv "github.com/pixie-sh/core-go/pkg/env"
)

type ProxyResponse = *goHttp.Response

type Proxy struct {
	Client
}

func (p Proxy) Forward(ctx context.Context, method string, fullURL string, body *bytes.Buffer, extraHeaders ...HeaderEntry) (ProxyResponse, error) {
	req, err := goHttp.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, errors.NewWithError(err, "error initializing request").WithErrorCode(errors.ErrorPerformingRequestErrorCode)
	}

	req = req.WithContext(ctx)
	for _, header := range extraHeaders {
		req.Header.Set(header.Key, header.Value)
	}

	userAgent := pixieEnv.EnvUserAgent()
	if userAgent == "" {
		userAgent = "Pixie-Sh"
	}

	req.Header.Set("User-Agent", fmt.Sprintf("%s (%s/%s)", userAgent, env.EnvScope(), env.EnvAppVersion()))
	res, err := p.client.Do(req)
	if err != nil {
		return nil, errors.NewWithError(err, "error performing rest request").WithErrorCode(errors.ErrorPerformingRequestErrorCode)
	}

	return res, nil
}

func NewProxy(ctx context.Context, cfg *ClientConfiguration) Proxy {
	return Proxy{NewClient(ctx, cfg)}
}

func DecompressBrotli(data []byte) ([]byte, error) {
	reader := brotli.NewReader(bytes.NewReader(data))
	decompressed, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return decompressed, nil
}

func DecompressGzip(compressedData []byte) ([]byte, error) {
	// Create a bytes buffer from the compressed data
	buf := bytes.NewBuffer(compressedData)
	// Create a new gzip Reader
	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer func(gzipReader *gzip.Reader) {
		_ = gzipReader.Close()
	}(gzipReader)

	// Decompress the data
	decompressedData, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		return nil, err
	}

	return decompressedData, nil
}
