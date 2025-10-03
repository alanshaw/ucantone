package receipt

import (
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/ipfs/go-cid"
)

// Option is an option configuring a UCAN receipt.
type Option func(cfg *receiptConfig)

type receiptConfig struct {
	exp   *ucan.UTCUnixTimestamp
	noexp bool
	nnc   []byte
	nonnc bool
	meta  ipld.Map[string, any]
	prf   []cid.Cid
}

// WithExpiration configures the expiration time in UTC seconds since Unix
// epoch.
func WithExpiration(exp ucan.UTCUnixTimestamp) Option {
	return func(cfg *receiptConfig) {
		cfg.exp = &exp
		cfg.noexp = false
	}
}

// WithNoExpiration configures the UCAN to never expire.
//
// WARNING: this will cause the delegation to be valid FOREVER, unless revoked.
func WithNoExpiration() Option {
	return func(cfg *receiptConfig) {
		cfg.exp = nil
		cfg.noexp = true
	}
}

// WithNonce configures the nonce value for the UCAN.
func WithNonce(nnc ucan.Nonce) Option {
	return func(cfg *receiptConfig) {
		cfg.nnc = nnc
	}
}

// WithNoNonce configures an empty nonce value for the UCAN.
func WithNoNonce() Option {
	return func(cfg *receiptConfig) {
		cfg.nonnc = true
	}
}

// WithMetadata configures the arbitrary metadata for the UCAN.
func WithMetadata(meta ipld.Map[string, any]) Option {
	return func(cfg *receiptConfig) {
		cfg.meta = meta
	}
}

// WithProof configures the proof(s) for the UCAN. If the `issuer` of this
// `Invocation` is not the resource owner / service provider, for the delegated
// capabilities, the `proofs` must contain valid `Proof`s containing
// delegations to the `issuer`.
func WithProofs(prf ...ucan.Link) Option {
	return func(cfg *receiptConfig) {
		cfg.prf = prf
	}
}
