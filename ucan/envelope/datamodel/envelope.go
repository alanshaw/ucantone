package datamodel

type EnvelopeModel[T any] struct {
	Signature  []byte
	SigPayload T
}
