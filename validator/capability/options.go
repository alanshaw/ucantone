package capability

import (
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
)

type capabilityConfig struct {
	pol policy.Policy
}

// Option is an option configuring a capability definition.
type Option func(cfg *capabilityConfig) error

func WithPolicy(pol ucan.Policy) Option {
	return func(cfg *capabilityConfig) error {
		pol, err := policy.New(pol.Statements()...)
		if err != nil {
			return err
		}
		cfg.pol = pol
		return nil
	}
}
