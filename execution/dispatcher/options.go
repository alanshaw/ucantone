package dispatcher

import (
	"time"

	"github.com/fil-forge/ucantone/validator"
)

// DefaultHandlerErrorReceiptTTL is the default TTL for receipts issued when a
// handler returns an error. Errors returned from handlers are unexpected and
// likely transient (permanent errors are set as error results on the
// response), so their receipts are short-lived assertions.
const DefaultHandlerErrorReceiptTTL = 30 * time.Second

// Option is an option configuring a UCAN executor.
type Option func(cfg *execConfig)

type execConfig struct {
	validationOpts         []validator.Option
	receiptTimestamps      bool
	handlerErrorReceiptTTL time.Duration
}

func WithValidationOptions(options ...validator.Option) Option {
	return func(cfg *execConfig) {
		cfg.validationOpts = append(cfg.validationOpts, options...)
	}
}

// WithReceiptTimestamps configures the dispatcher to issue receipts with
// issuance timestamps or not.
func WithReceiptTimestamps(enabled bool) Option {
	return func(cfg *execConfig) {
		cfg.receiptTimestamps = enabled
	}
}

// WithHandlerErrorReceiptTTL configures the TTL for receipts issued when a
// handler returns an error. A zero or negative value disables the expiration,
// making the receipt a permanent assertion. Defaults to
// [DefaultHandlerErrorReceiptTTL].
func WithHandlerErrorReceiptTTL(ttl time.Duration) Option {
	return func(cfg *execConfig) {
		cfg.handlerErrorReceiptTTL = ttl
	}
}
