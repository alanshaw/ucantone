package signer

import (
	"fmt"
	"strings"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/principal"
	"github.com/alanshaw/ucantone/principal/verifier"
)

type Unwrapper interface {
	// Unwrap returns the unwrapped did:key of this signer.
	Unwrap() principal.Signer
}

type WrappedSigner struct {
	key      principal.Signer
	verifier principal.Verifier
}

func (w *WrappedSigner) Code() uint64 {
	return w.key.Code()
}

func (w *WrappedSigner) DID() did.DID {
	return w.verifier.DID()
}

func (w *WrappedSigner) Bytes() []byte {
	return w.key.Bytes()
}

func (w *WrappedSigner) Raw() []byte {
	return w.key.Raw()
}

func (w *WrappedSigner) Sign(msg []byte) []byte {
	return w.key.Sign(msg)
}

func (w *WrappedSigner) SignatureCode() uint64 {
	return w.key.SignatureCode()
}

func (w *WrappedSigner) Unwrap() principal.Signer {
	return w.key
}

func (w *WrappedSigner) Verifier() principal.Verifier {
	return w.verifier
}

// Wrap the key of this signer into a signer with a different DID. This is
// primarily used to wrap a did:key signer with a signer that has a DID of
// a different method.
func Wrap(key principal.Signer, id did.DID) (*WrappedSigner, error) {
	if !strings.HasPrefix(key.DID().String(), "did:key:") {
		return nil, fmt.Errorf("verifier is not a did:key")
	}
	vrf, err := verifier.Wrap(key.Verifier(), id)
	if err != nil {
		return nil, err
	}
	return &WrappedSigner{key, vrf}, nil
}
