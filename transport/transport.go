package transport

import (
	"context"

	"github.com/alanshaw/ucantone/ucan"
)

// Contexter is an interface that provides a context.
type Contexter interface {
	Context() context.Context
}

type InboundCodec[Req Contexter, Res any] interface {
	Decode(Req) (ucan.Container, error)
	Encode(ucan.Container) (Res, error)
}

type OutboundCodec[Req Contexter, Res any] interface {
	Encode(ucan.Container) (Req, error)
	Decode(Res) (ucan.Container, error)
}

type RoundTripper[Req Contexter, Res any] interface {
	RoundTrip(Req) (Res, error)
}
