package client

import (
	"net/http"
	"net/url"

	"github.com/alanshaw/ucantone/execution"
	"github.com/alanshaw/ucantone/transport"
)

type HTTPClient struct {
	*Client[*http.Request, *http.Response]
}

func NewHTTP(serviceURL *url.URL, options ...HTTPOption) (*HTTPClient, error) {
	cfg := httpClientConfig{
		codec:  transport.DefaultHTTPOutboundCodec,
		client: http.DefaultClient,
	}
	for _, opt := range options {
		opt(&cfg)
	}
	return &HTTPClient{
		Client: New(&httpTransport{cfg.client, serviceURL}, cfg.codec),
	}, nil
}

func (c *HTTPClient) Execute(execRequest execution.Request) (execution.Response, error) {
	return c.Client.Execute(execRequest)
}

type httpTransport struct {
	client *http.Client
	url    *url.URL
}

func (t *httpTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Method = http.MethodPost
	r.URL = t.url
	return t.client.Do(r)
}
