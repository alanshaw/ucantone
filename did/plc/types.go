package plc

import (
	"fmt"
	"maps"
	"slices"

	"github.com/fil-forge/ucantone/did"
	cid "github.com/ipfs/go-cid"
)

const (
	OperationType = "plc_operation"
	TombstoneType = "plc_tombstone"
)

var ErrMissingRotationKeys = fmt.Errorf("at least one rotation key is required")

// Signer is an entity that can sign a payload.
type Signer interface {
	// Sign takes a byte encoded message and produces a verifiable signature.
	Sign(msg []byte) []byte
}

// Verifier represents an entity that can verify signatures.
type Verifier interface {
	// Verify takes a byte encoded message and verifies that it is signed by
	// corresponding signer.
	Verify(msg []byte, sig []byte) bool
}

// Operation represents a PLC operation that can be used to create or update a
// PLC DID.
type Operation struct {
	Type                string             `cborgen:"type,const=plc_operation" dagjsongen:"type,const=plc_operation"`
	VerificationMethods map[string]did.DID `cborgen:"verificationMethods" dagjsongen:"verificationMethods"`
	RotationKeys        []did.DID          `cborgen:"rotationKeys" dagjsongen:"rotationKeys"`
	AlsoKnownAs         []string           `cborgen:"alsoKnownAs" dagjsongen:"alsoKnownAs"`
	Services            map[string]Service `cborgen:"services" dagjsongen:"services"`
	// String encoded CID of the previous operation in the chain, if any. If this
	// is the first operation in the chain, this field is null.
	Previous *string `cborgen:"prev" dagjsongen:"prev"`
}

type SignedOperation struct {
	Type                string             `cborgen:"type,const=plc_operation" dagjsongen:"type,const=plc_operation"`
	VerificationMethods map[string]did.DID `cborgen:"verificationMethods" dagjsongen:"verificationMethods"`
	RotationKeys        []did.DID          `cborgen:"rotationKeys" dagjsongen:"rotationKeys"`
	AlsoKnownAs         []string           `cborgen:"alsoKnownAs" dagjsongen:"alsoKnownAs"`
	Services            map[string]Service `cborgen:"services" dagjsongen:"services"`
	// String encoded CID of the previous operation in the chain, if any. If this
	// is the first operation in the chain, this field is null.
	Previous  *string `cborgen:"prev" dagjsongen:"prev"`
	Signature string  `cborgen:"sig" dagjsongen:"sig"`
}

type Service struct {
	Type     string `cborgen:"type" dagjsongen:"type"`
	Endpoint string `cborgen:"endpoint" dagjsongen:"endpoint"`
}

type Tombstone struct {
	Type     string `cborgen:"type,const=plc_tombstone" dagjsongen:"type,const=plc_tombstone"`
	Previous string `cborgen:"prev" dagjsongen:"prev"`
}

type SignedTombstone struct {
	Type      string `cborgen:"type,const=plc_tombstone" dagjsongen:"type,const=plc_tombstone"`
	Previous  string `cborgen:"prev" dagjsongen:"prev"`
	Signature string `cborgen:"sig" dagjsongen:"sig"`
}

type opConfig struct {
	verificationMethods map[string]did.DID
	rotationKeys        []did.DID
	alsoKnownAs         []string
	services            map[string]Service
}

type OperationOption func(*opConfig)

// WithVerificationMethod adds a verification method to the PLC operation.
func WithVerificationMethods(methods map[string]did.DID) OperationOption {
	return func(c *opConfig) {
		if c.verificationMethods == nil {
			c.verificationMethods = make(map[string]did.DID, len(methods))
		}
		maps.Copy(c.verificationMethods, methods)
	}
}

// WithoutVerificationMethods removes the given verification methods from the
// PLC operation.
func WithoutVerificationMethods(methods map[string]did.DID) OperationOption {
	return func(c *opConfig) {
		for m := range methods {
			delete(c.verificationMethods, m)
		}
	}
}

// WithRotationKeys adds rotation keys to the PLC operation.
func WithRotationKeys(keys ...did.DID) OperationOption {
	return func(c *opConfig) {
		c.rotationKeys = append(c.rotationKeys, keys...)
	}
}

// WithoutRotationKeys removes the given rotation keys from the PLC operation.
func WithoutRotationKeys(keys ...did.DID) OperationOption {
	return func(c *opConfig) {
		c.rotationKeys = slices.DeleteFunc(c.rotationKeys, func(k did.DID) bool {
			return slices.Contains(keys, k)
		})
	}
}

// WithAlsoKnownAs adds also known as entries to the PLC operation.
func WithAlsoKnownAs(alsoKnownAs ...string) OperationOption {
	return func(c *opConfig) {
		c.alsoKnownAs = append(c.alsoKnownAs, alsoKnownAs...)
	}
}

// WithoutAlsoKnownAs removes the given also known as entries from the PLC
// operation.
func WithoutAlsoKnownAs(alsoKnownAs ...string) OperationOption {
	return func(c *opConfig) {
		c.alsoKnownAs = slices.DeleteFunc(c.alsoKnownAs, func(a string) bool {
			return slices.Contains(alsoKnownAs, a)
		})
	}
}

// WithServices adds services to the PLC operation.
func WithServices(services map[string]Service) OperationOption {
	return func(c *opConfig) {
		if c.services == nil {
			c.services = make(map[string]Service, len(services))
		}
		maps.Copy(c.services, services)
	}
}

// WithoutServices removes the given services from the PLC operation.
func WithoutServices(services map[string]Service) OperationOption {
	return func(c *opConfig) {
		for s := range services {
			delete(c.services, s)
		}
	}
}

// NewOperation creates a new PLC operation with the given previous operation
// CID and options.
func NewOperation(prev *cid.Cid, options ...OperationOption) (*Operation, error) {
	cfg := opConfig{
		verificationMethods: map[string]did.DID{},
		services:            map[string]Service{},
	}
	for _, option := range options {
		option(&cfg)
	}
	if len(cfg.rotationKeys) == 0 {
		return nil, ErrMissingRotationKeys
	}
	var prevStr *string
	if prev != nil {
		s := prev.String()
		prevStr = &s
	}
	return &Operation{
		Type:                OperationType,
		VerificationMethods: cfg.verificationMethods,
		RotationKeys:        cfg.rotationKeys,
		AlsoKnownAs:         cfg.alsoKnownAs,
		Services:            cfg.services,
		Previous:            prevStr,
	}, nil
}

// NewTombstone creates a new PLC tombstone with the given previous operation
// CID. The tombstone indicates that the DID has been deactivated and should no
// longer be used.
func NewTombstone(prev cid.Cid) *Tombstone {
	return &Tombstone{
		Type:     TombstoneType,
		Previous: prev.String(),
	}
}
