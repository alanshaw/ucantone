package server

import (
	"net/http"

	"github.com/alanshaw/ucantone/transport"
	"github.com/alanshaw/ucantone/validator"
)

// Option is an option configuring a UCAN executor.
type Option func(cfg *serverConfig)

type serverConfig struct {
	codec          transport.InboundCodec[*http.Request, *http.Response]
	validationOpts []validator.Option
}

func WithCodec(codec transport.InboundCodec[*http.Request, *http.Response]) Option {
	return func(cfg *serverConfig) {
		cfg.codec = codec
	}
}

func WithValidationOptions(options ...validator.Option) Option {
	return func(cfg *serverConfig) {
		cfg.validationOpts = append(cfg.validationOpts, options...)
	}
}
