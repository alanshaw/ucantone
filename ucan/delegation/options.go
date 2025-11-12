package delegation

import (
	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
)

// Option is an option configuring a UCAN invocation.
type Option func(cfg *delegationConfig) error

type delegationConfig struct {
	sub       *did.DID
	powerline bool
	exp       *ucan.UTCUnixTimestamp
	nbf       *ucan.UTCUnixTimestamp
	noexp     bool
	nnc       []byte
	nonnc     bool
	meta      ipld.Map[string, ipld.Any]
	pol       policy.Policy
}

// WithSubject configures the DID of the subject of the delegation chain.
func WithSubject(sub ucan.Principal) Option {
	return func(cfg *delegationConfig) error {
		if sub == nil {
			cfg.sub = nil
		} else {
			sub := sub.DID()
			cfg.sub = &sub
		}
		return nil
	}
}

// WithPowerline configures the delegation powerline. Setting powerline to true
// allows the delegation subject to be unset.
//
// "Powerline" is a pattern for automatically delegating all future delegations
// to another agent regardless of Subject.
//
// https://github.com/ucan-wg/delegation/blob/main/README.md#powerline
func WithPowerline(on bool) Option {
	return func(cfg *delegationConfig) error {
		cfg.powerline = on
		return nil
	}
}

// WithExpiration configures the expiration time in UTC seconds since Unix
// epoch.
func WithExpiration(exp ucan.UTCUnixTimestamp) Option {
	return func(cfg *delegationConfig) error {
		cfg.exp = &exp
		cfg.noexp = false
		return nil
	}
}

// WithNoExpiration configures the UCAN to never expire.
//
// WARNING: this will cause the delegation to be valid FOREVER, unless revoked.
func WithNoExpiration() Option {
	return func(cfg *delegationConfig) error {
		cfg.exp = nil
		cfg.noexp = true
		return nil
	}
}

// WithNonce configures the nonce value for the UCAN.
func WithNonce(nnc ucan.Nonce) Option {
	return func(cfg *delegationConfig) error {
		cfg.nnc = nnc
		return nil
	}
}

// WithNoNonce configures an empty nonce value for the UCAN.
func WithNoNonce() Option {
	return func(cfg *delegationConfig) error {
		cfg.nonnc = true
		return nil
	}
}

// WithNotBefore configures the time in UTC seconds since Unix epoch that the
// delegation becomes valid.
func WithNotBefore(nbf ucan.UTCUnixTimestamp) Option {
	return func(cfg *delegationConfig) error {
		cfg.nbf = &nbf
		return nil
	}
}

// WithMetadata configures the arbitrary metadata for the UCAN.
func WithMetadata(meta ipld.Map[string, ipld.Any]) Option {
	return func(cfg *delegationConfig) error {
		cfg.meta = meta
		return nil
	}
}

func WithPolicy(pol ucan.Policy) Option {
	return func(cfg *delegationConfig) error {
		pol, err := policy.New(pol.Statements()...)
		if err != nil {
			return err
		}
		cfg.pol = pol
		return nil
	}
}
