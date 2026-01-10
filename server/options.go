package server

import (
	"net/http"

	"github.com/alanshaw/ucantone/transport"
	"github.com/alanshaw/ucantone/validator"
)

// HTTPOption is an option configuring a UCAN HTTP server.
type HTTPOption func(cfg *httpServerConfig)

type httpServerConfig struct {
	codec          transport.InboundCodec[*http.Request, *http.Response]
	validationOpts []validator.Option
	listeners      []EventListener
}

func WithHTTPCodec(codec transport.InboundCodec[*http.Request, *http.Response]) HTTPOption {
	return func(cfg *httpServerConfig) {
		cfg.codec = codec
	}
}

func WithValidationOptions(options ...validator.Option) HTTPOption {
	return func(cfg *httpServerConfig) {
		cfg.validationOpts = append(cfg.validationOpts, options...)
	}
}

// WithEventListener adds an event listener to the HTTP server for monitoring
// requests and responses.
func WithEventListener(listener EventListener) HTTPOption {
	return func(cfg *httpServerConfig) {
		cfg.listeners = append(cfg.listeners, listener)
	}
}
