//go:build !codegen

package plc

import (
	"bytes"
	"fmt"
	"maps"
	"slices"

	"github.com/fil-forge/ucantone/did"
	cid "github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

// NewOperationFromPrevious creates a new PLC operation that updates the given
// previous operation with the provided options. The new operation will have the
// previous verification methods, rotation keys, also known as, and services as
// the previous operation, merged with the values passed in the options.
func NewOperationFromPrevious(prev *SignedOperation, options ...OperationOption) (*Operation, error) {
	cfg := opConfig{
		verificationMethods: map[string]did.DID{},
		services:            map[string]Service{},
	}
	if len(prev.RotationKeys) != 0 {
		cfg.rotationKeys = slices.Clone(prev.RotationKeys)
	}
	if len(prev.AlsoKnownAs) != 0 {
		cfg.alsoKnownAs = slices.Clone(prev.AlsoKnownAs)
	}
	maps.Copy(cfg.verificationMethods, prev.VerificationMethods)
	maps.Copy(cfg.services, prev.Services)

	for _, option := range options {
		option(&cfg)
	}

	if len(cfg.rotationKeys) == 0 {
		return nil, ErrMissingRotationKeys
	}

	prevLink, err := SumOperation(prev)
	if err != nil {
		return nil, err
	}
	prevLinkStr := prevLink.String()
	return &Operation{
		Type:                OperationType,
		VerificationMethods: cfg.verificationMethods,
		RotationKeys:        cfg.rotationKeys,
		AlsoKnownAs:         cfg.alsoKnownAs,
		Services:            cfg.services,
		Previous:            &prevLinkStr,
	}, nil
}

// SumOperation computes the CID of a signed operation, as used to link the next
// operation in the chain to its predecessor.
func SumOperation(op *SignedOperation) (cid.Cid, error) {
	var opBytes bytes.Buffer
	if err := op.MarshalCBOR(&opBytes); err != nil {
		return cid.Undef, err
	}
	link, err := cid.V1Builder{
		Codec:  cid.DagCBOR,
		MhType: multihash.SHA2_256,
	}.Sum(opBytes.Bytes())
	if err != nil {
		return cid.Undef, fmt.Errorf("hashing previous operation: %w", err)
	}
	return link, nil
}

// NewTombstoneFromPrevious creates a new PLC tombstone that deactivates
// the DID, linking to the given previous operation by its computed CID. It is a
// convenience over [NewTombstone] for the common case where you have fetched the
// last signed operation (e.g. via DirectoryClient.Last) rather than its CID.
func NewTombstoneFromPrevious(prev *SignedOperation) (*Tombstone, error) {
	prevLink, err := SumOperation(prev)
	if err != nil {
		return nil, err
	}
	return &Tombstone{
		Type:     TombstoneType,
		Previous: prevLink.String(),
	}, nil
}
