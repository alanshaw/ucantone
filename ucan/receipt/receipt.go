package receipt

import (
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/invocation"
)

// FIXME: receipt _is an_ invocation but envelope tag is different
type Receipt = invocation.Invocation

func Encode(receipt ucan.Receipt) ([]byte, error) {
	return invocation.Encode(receipt)
}

func Decode(input []byte) (*Receipt, error) {
	return invocation.Decode(input)
}
