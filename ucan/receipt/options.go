package receipt

import (
	"github.com/fil-forge/ucantone/ipld"
	"github.com/fil-forge/ucantone/ucan"
	"github.com/fil-forge/ucantone/ucan/invocation"
)

// Option configures a UCAN receipt.
//
// A receipt is encoded on the wire as a /ucan/assert/receipt invocation (see
// ucan-wg/receipt#1), so each receipt Option translates to an invocation
// Option internally. Only the options that are meaningful for a receipt are
// exposed here; proof options will be added once delegated receipt issuance
// is supported.
type Option func(cfg *receiptConfig)

type receiptConfig struct {
	invOpts []invocation.Option
}

// WithNonce configures the nonce value for the receipt.
func WithNonce(nnc []byte) Option {
	return func(cfg *receiptConfig) {
		cfg.invOpts = append(cfg.invOpts, invocation.WithNonce(nnc))
	}
}

// WithNoNonce configures an empty nonce value for the receipt.
func WithNoNonce() Option {
	return func(cfg *receiptConfig) {
		cfg.invOpts = append(cfg.invOpts, invocation.WithNoNonce())
	}
}

// WithExpiration sets the time until which the executor commits to uphold
// the asserted result, in seconds since the Unix epoch. Use it when the
// result is impermanent (e.g. it references state that may be garbage
// collected). By default a receipt never expires (exp: null), which is
// appropriate for permanent facts such as the result of a pure computation.
func WithExpiration(exp ucan.UnixTimestamp) Option {
	return func(cfg *receiptConfig) {
		cfg.invOpts = append(cfg.invOpts, invocation.WithExpiration(exp))
	}
}

// WithIssuedAt sets the time at which the receipt was issued, in seconds
// since the Unix epoch.
func WithIssuedAt(iat ucan.UnixTimestamp) Option {
	return func(cfg *receiptConfig) {
		cfg.invOpts = append(cfg.invOpts, invocation.WithIssuedAt(iat))
	}
}

// WithMetadata configures arbitrary metadata for the receipt.
func WithMetadata(meta ipld.Map) Option {
	return func(cfg *receiptConfig) {
		cfg.invOpts = append(cfg.invOpts, invocation.WithMetadata(meta))
	}
}
