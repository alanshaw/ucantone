package receipt

import "github.com/alanshaw/ucantone/ucan/invocation"

type Option = invocation.Option

var (
	WithAudience     = invocation.WithAudience
	WithExpiration   = invocation.WithExpiration
	WithNoExpiration = invocation.WithNoExpiration
	WithNonce        = invocation.WithNonce
	WithNoNonce      = invocation.WithNoNonce
	WithMetadata     = invocation.WithMetadata
	WithProofs       = invocation.WithProofs
	WithCause        = invocation.WithCause
)
