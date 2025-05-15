package ihttp

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"sync/atomic"

	"github.com/google/logger"
)

var _ http.RoundTripper = (*CustomTransport)(nil)
var _ http.RoundTripper = (*LoggingTransport)(nil)

// Custom transport to add headers to all requests.
type CustomTransport struct {
	Next         http.RoundTripper
	UserAgent    string
	ExtraHeaders map[string]string
}

// RoundTrip implements the http.RoundTripper interface.
func (t *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original.
	req = req.Clone(req.Context())
	if t.UserAgent != "" {
		req.Header.Set("User-Agent", t.UserAgent)
	}
	for key, value := range t.ExtraHeaders {
		req.Header.Set(key, value)
	}
	// Use the base transport to perform the actual request.
	return t.Next.RoundTrip(req)
}

// Logs all requests.
type LoggingTransport struct {
	Next http.RoundTripper
	i    atomic.Int32
}

func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	i := t.i.Add(1) - 1
	b, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		logger.Warningf("http[%d]: request error, %v", i, err)
		return nil, err
	}
	logger.Infof("http[%d]: request\n%s", i, string(prefix(b, "  ")))

	res, err := t.Next.RoundTrip(req)
	b, err = httputil.DumpResponse(res, false)
	if err != nil {
		logger.Warningf("http[%d]: response error, %v", i, err)
		return res, err
	}
	logger.Infof("http[%d]: response\n%s", i, string(prefix(b, "  ")))

	return res, err
}

func prefix(data []byte, prefix string) []byte {
	data = bytes.TrimRight(data, "\r\n")
	if len(data) == 0 {
		return data
	}
	nl := []byte("\r\n")
	nlprefix := []byte("\r\n" + prefix)
	ndata := bytes.ReplaceAll(data, nl, nlprefix)
	return append([]byte(prefix), ndata...)
}
