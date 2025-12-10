package executor

import (
	"github.com/alanshaw/ucantone/validator"
)

// Option is an option configuring a UCAN executor.
type Option func(cfg *execConfig)

type execConfig struct {
	validationOpts []validator.Option
}

func WithValidationOptions(options ...validator.Option) Option {
	return func(cfg *execConfig) {
		cfg.validationOpts = append(cfg.validationOpts, options...)
	}
}
