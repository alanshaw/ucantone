package invocation

import (
	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/ipfs/go-cid"
)

// Option is an option configuring a UCAN invocation.
type Option func(cfg *invocationConfig) error

type invocationConfig struct {
	aud   *did.DID
	exp   *ucan.UTCUnixTimestamp
	noexp bool
	nnc   []byte
	nonnc bool
	meta  dagcbor.CBORMarshalable
	prf   []cid.Cid
	cause *cid.Cid
}

// WithAudience configures the DID of the intended Executor if different from
// the Subject.
func WithAudience(aud did.DID) Option {
	return func(cfg *invocationConfig) error {
		cfg.aud = &aud
		return nil
	}
}

// WithExpiration configures the expiration time in UTC seconds since Unix
// epoch.
func WithExpiration(exp ucan.UTCUnixTimestamp) Option {
	return func(cfg *invocationConfig) error {
		cfg.exp = &exp
		cfg.noexp = false
		return nil
	}
}

// WithNoExpiration configures the UCAN to never expire.
//
// WARNING: this will cause the delegation to be valid FOREVER, unless revoked.
func WithNoExpiration() Option {
	return func(cfg *invocationConfig) error {
		cfg.exp = nil
		cfg.noexp = true
		return nil
	}
}

// WithNonce configures the nonce value for the UCAN.
func WithNonce(nnc ucan.Nonce) Option {
	return func(cfg *invocationConfig) error {
		cfg.nnc = nnc
		return nil
	}
}

// WithNoNonce configures an empty nonce value for the UCAN.
func WithNoNonce() Option {
	return func(cfg *invocationConfig) error {
		cfg.nonnc = true
		return nil
	}
}

// WithMetadata configures the arbitrary metadata for the UCAN.
func WithMetadata(meta dagcbor.CBORMarshalable) Option {
	return func(cfg *invocationConfig) error {
		cfg.meta = meta
		return nil
	}
}

// WithProof configures the proof(s) for the UCAN. If the `issuer` of this
// `Invocation` is not the resource owner / service provider, for the delegated
// capabilities, the `proofs` must contain valid `Proof`s containing
// delegations to the `issuer`.
func WithProofs(prf ...ucan.Link) Option {
	return func(cfg *invocationConfig) error {
		cfg.prf = prf
		return nil
	}
}

// WithCause configures the CID of the receipt that enqueued the task.
func WithCause(cause ucan.Link) Option {
	return func(cfg *invocationConfig) error {
		cfg.cause = &cause
		return nil
	}
}
