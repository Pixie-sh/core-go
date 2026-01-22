package rest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	goHttp "net/http"
	"net/url"
	"time"

	pixieEnv "github.com/pixie-sh/core-go/pkg/env"

	"github.com/pixie-sh/errors-go"
	"github.com/pixie-sh/logger-go/caller"
	"github.com/pixie-sh/logger-go/env"
	"github.com/pixie-sh/logger-go/logger"

	"github.com/pixie-sh/core-go/pkg/comm/http"
	pixiecontext "github.com/pixie-sh/core-go/pkg/context"
	"github.com/pixie-sh/core-go/pkg/types"
)

// IClient basic rest client functionality
type IClient interface {
	Do(ctx context.Context, method http.Method, fullURL string, body io.Reader, extraHeaders ...HeaderEntry) (io.ReadCloser, *goHttp.Response, error)
	DoRAW(ctx context.Context, method http.Method, fullURL string, body io.Reader, extraHeaders ...HeaderEntry) ([]byte, error)
	DoJSON(ctx context.Context, method http.Method, fullURL string, body io.Reader, result interface{}, extraHeaders ...HeaderEntry) error
}

// HeaderEntry header entry
type HeaderEntry struct {
	Key   string
	Value string
}

// Client implements IClient
type Client struct {
	headerKey string
	apiKeys   map[string]string
	client    *goHttp.Client
}

// NewClient receives the header key to be used for ms authorization,
// a map with host and token to be used when performing request to those hosts
// timeout for rest request hanging. returns a Client ptr that implements IClient
//
// Disable HTTP/2 to avoid HPACK encoder panic under high concurrency.
// See: https://github.com/golang/go/issues/47882
func NewClient(_ context.Context, cfg *ClientConfiguration) Client {
	var client *goHttp.Client
	if cfg.GoClient != nil {
		client = cfg.GoClient
	} else {
		client = &goHttp.Client{
			Timeout: time.Millisecond * time.Duration(cfg.Timeout),
			Transport: &goHttp.Transport{
				ForceAttemptHTTP2: false,
				TLSNextProto:      make(map[string]func(authority string, c *tls.Conn) goHttp.RoundTripper),
			},
		}
	}

	return Client{
		headerKey: cfg.HeaderAPIKey,
		apiKeys:   cfg.APIKeys,
		client:    client,
	}
}

// NewNakedClient receives the header key to be used for ms authorization,
// a map with host and token to be used when performing request to those hosts
// timeout for rest request hanging. returns a Client ptr that implements IClient
// Disable HTTP/2 to avoid HPACK encoder panic under high concurrency.
// See: https://github.com/golang/go/issues/47882
func NewNakedClient(timeout time.Duration) *Client {
	return &Client{
		apiKeys: make(map[string]string),
		client: &goHttp.Client{
			Timeout: timeout,
			Transport: &goHttp.Transport{
				ForceAttemptHTTP2: false,
				TLSNextProto:      make(map[string]func(authority string, c *tls.Conn) goHttp.RoundTripper),
			},
		},
	}
}

func NewNakedClientWithGoClient(_ context.Context, client *goHttp.Client) *Client {
	return &Client{
		apiKeys: make(map[string]string),
		client:  client,
	}
}

func (c *Client) Do(ctx context.Context, method http.Method, fullURL string, body io.Reader, extraHeaders ...HeaderEntry) (io.ReadCloser, *goHttp.Response, error) {
	req, err := goHttp.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, nil, errors.NewWithError(err, "error initializing request").WithErrorCode(errors.ErrorPerformingRequestErrorCode)
	}

	userAgent := pixieEnv.EnvUserAgent()
	if userAgent == "" {
		userAgent = "Pixie-Sh"
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("User-Agent", fmt.Sprintf("%s (%s/%s)", userAgent, env.EnvScope(), env.EnvAppVersion()))
	req.Header.Set("Accept", "application/json; charset=utf-8")
	for _, header := range extraHeaders {
		req.Header.Set(header.Key, header.Value)
	}

	apiKey := getHostAPIKey(fullURL, c.apiKeys)
	if apiKey != "" {
		req.Header.Set(c.headerKey, apiKey)
	}

	var curl []byte
	if env.IsDebugActive() {
		curl, _ = printCurlFormat(req)
	}

	pixiecontext.GetCtxLogger(ctx).
		With("request", goHttpRequestToPrintable(req)).
		With("request.curl", curl).
		With("threeHopsCaller", caller.NewCaller(caller.ThreeHopsCallerDepth)).
		With("fourHopsCaller", caller.NewCaller(caller.FourHopsCallerDepth)).
		Debug("executing request to %s", fullURL)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, res, errors.NewWithError(err, "error performing rest request").WithErrorCode(errors.ErrorPerformingRequestErrorCode)
	}

	if res.StatusCode < goHttp.StatusOK || res.StatusCode >= goHttp.StatusBadRequest {
		ec := errors.ErrorPerformingRequestErrorCode

		ec.HTTPError = res.StatusCode
		return res.Body, res, errors.New("rest response %d on %s %s", res.StatusCode, req.Method, req.URL.Path).WithErrorCode(ec)
	}

	return res.Body, res, nil
}

func printCurlFormat(req *goHttp.Request) ([]byte, error) {
	var curlCmd bytes.Buffer

	// Start with the curl method and URL
	curlCmd.WriteString(fmt.Sprintf("curl -X %s '%s'", req.Method, req.URL.String()))

	// Include headers
	for name, values := range req.Header {
		for _, value := range values {
			if len(value) == 0 {
				continue
			}
			curlCmd.WriteString(fmt.Sprintf(" -H '%s: %s'", name, value))
		}
	}

	// Include body if there's any
	if req.Body != nil {
		bodyBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		// Make sure to reset the request body
		req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

		if len(bodyBytes) > 0 {
			curlCmd.WriteString(fmt.Sprintf(" -d '%s'", string(bodyBytes)))
		}
	}

	return curlCmd.Bytes(), nil
}

// DoRAW execute rest call and return response bytes
func (c *Client) DoRAW(ctx context.Context, method http.Method, fullURL string, body io.Reader, extraHeaders ...HeaderEntry) ([]byte, error) {
	resp, _, err := c.Do(ctx, method, fullURL, body, extraHeaders...)
	if resp != nil {
		defer func(resp io.ReadCloser) {
			_ = resp.Close()
		}(resp)

		bytes, bErr := io.ReadAll(resp)
		if bErr != nil {
			pixiecontext.GetCtxLogger(ctx).Debug("rest client read %d bytes; error reading bytes? %s", len(bytes), bErr)
			return []byte{}, errors.New("unable to read response bytes; %s", bErr)
		}

		pixiecontext.GetCtxLogger(ctx).With("raw_response", bytes).Debug("rest client read %d bytes", len(bytes))
		if err != nil {
			return bytes, err //return error body
		}

		return bytes, nil
	}

	return []byte{}, err
}

// DoJSON execute Do request but unmarshall the body to struct. unmarshall for json only
// if responseStruct is nil, no unmarshall is performed
// if responseStruct not nil, must be a pointer, so the unmarshal can fill it up
func (c *Client) DoJSON(ctx context.Context, method http.Method, fullURL string, body io.Reader, responseStruct interface{}, extraHeaders ...HeaderEntry) error {
	ctxLogger := pixiecontext.GetCtxLogger(ctx)
	if responseStruct != nil && !types.IsPointer(responseStruct) {
		return errors.New("response struct object must be pointer").WithErrorCode(errors.ErrorPerformingRequestErrorCode)
	}

	blob, err := c.DoRAW(ctx, method, fullURL, body, extraHeaders...)
	ctxLogger.With("response", string(blob)).Debug("response from rest call")
	if err != nil {
		return errors.New("unable to execute rest request: %s", err.Error()).WithErrorCode(errors.ErrorPerformingRequestErrorCode)
	}

	// response parsing required
	if responseStruct != nil {
		err = json.Unmarshal(blob, responseStruct)
		if err != nil {
			return errors.New("%s", err.Error()).WithErrorCode(errors.ErrorUnmarshallBodyErrorCode)
		}
	}

	// no error wo hoo
	return nil
}

func getHostAPIKey(fullURL string, keys map[string]string) string {
	if len(keys) == 0 {
		return ""
	}

	host, _ := url.Parse(fullURL)
	key, ok := keys[host.Host]
	if !ok {
		logger.Log("no api key for host %s", host.Host)
		return ""
	}

	return key
}
