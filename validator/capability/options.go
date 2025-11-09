package capability

import (
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
)

type capabilityConfig struct {
	pol policy.Policy
}

// Option is an option configuring a capability definition.
type Option func(cfg *capabilityConfig)

func WithPolicy(statements ...ucan.Statement) Option {
	return func(cfg *capabilityConfig) {
		cfg.pol = policy.New(statements...)
	}
}
