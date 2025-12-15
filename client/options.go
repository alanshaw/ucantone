package client

import (
	"net/http"

	"github.com/alanshaw/ucantone/transport"
)

type httpClientConfig struct {
	client *http.Client
	codec  transport.OutboundCodec[*http.Request, *http.Response]
}

type HTTPOption func(*httpClientConfig)

func WithHTTPClient(client *http.Client) HTTPOption {
	return func(cfg *httpClientConfig) {
		cfg.client = client
	}
}

func WithHTTPCodec(codec transport.OutboundCodec[*http.Request, *http.Response]) HTTPOption {
	return func(cfg *httpClientConfig) {
		cfg.codec = codec
	}
}
