package rest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
)

// HeaderEntryToQueryParamString Converts a list of HeaderEntries to a query param string.
func HeaderEntryToQueryParamString(params ...HeaderEntry) string {
	var list []string

	for _, value := range params {
		list = append(list, value.Key+"="+value.Value)
	}

	return strings.Join(list, "&")
}

// AppendQueryParamsToURL Appends query params to a given URL
func AppendQueryParamsToURL(url string, params ...HeaderEntry) string {
	urlBuild := url

	if len(params) > 0 {
		urlBuild += "?" + HeaderEntryToQueryParamString(params...)
	}
	return urlBuild
}

type printableRequest struct {
	Method           string              `json:"method"`
	URL              string              `json:"url"`
	Proto            string              `json:"proto"`
	Headers          map[string][]string `json:"headers"`
	ContentLength    int64               `json:"content_length"`
	TransferEncoding []string            `json:"transfer_encoding"`
	Close            bool                `json:"close"`
	Host             string              `json:"host"`
	RemoteAddr       string              `json:"remote_addr"`
	RequestURI       string              `json:"request_uri"`
	Body             []byte              `json:"body"`
}

func goHttpRequestToPrintable(req *http.Request) printableRequest {
	if req == nil {
		return printableRequest{
			URL: "<request nil>",
		}
	}

	pr := printableRequest{
		Method:           req.Method,
		URL:              req.URL.String(),
		Proto:            req.Proto,
		Headers:          req.Header,
		ContentLength:    req.ContentLength,
		TransferEncoding: req.TransferEncoding,
		Close:            req.Close,
		Host:             req.Host,
		RemoteAddr:       req.RemoteAddr,
		RequestURI:       req.RequestURI,
	}

	var err error
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return pr
		}

		req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	pr.Body = bodyBytes
	return pr
}
