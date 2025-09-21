package datamodel

type EnvelopeModel[T any] struct {
	// A signature by the Payload's iss over the SigPayload field.
	Signature []byte
	// The content that was signed.
	SigPayload T
}
