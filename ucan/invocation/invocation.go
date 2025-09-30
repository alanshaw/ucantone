package invocation

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/ucan"
	cmd "github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/crypto/signature"
	idm "github.com/alanshaw/ucantone/ucan/invocation/datamodel"
	"github.com/alanshaw/ucantone/ucan/nonce"
	"github.com/alanshaw/ucantone/varsig"
	"github.com/alanshaw/ucantone/varsig/algorithm/ed25519"
	"github.com/alanshaw/ucantone/varsig/common"
	cid "github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
)

// NoArguments can be used to issue an invocation with no arguments.
var NoArguments ipld.Map[string, any]

func init() {
	empty, err := datamodel.NewMap()
	if err != nil {
		panic(err)
	}
	NoArguments = empty
}

// UCAN Invocation defines a format for expressing the intention to execute
// delegated UCAN capabilities, and the attested receipts from an execution.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md
type Invocation struct {
	sig   *signature.Signature
	model *idm.EnvelopeModel
	args  *datamodel.Map
	meta  *datamodel.Map
}

// Parameters expected by the command.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#arguments
func (inv *Invocation) Arguments() ipld.Map[string, any] {
	return inv.args
}

// The DID of the intended Executor if different from the Subject.
//
// WARNING: May be nil.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#issuer--audience
func (inv *Invocation) Audience() ucan.Principal {
	if inv.model.SigPayload.TokenPayload1_0_0_rc1.Aud == nil {
		return nil
	}
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Aud
}

// A provenance claim describing which receipt requested it.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#cause
func (inv *Invocation) Cause() *cid.Cid {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Cause
}

// The command to invoke.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#command
func (inv *Invocation) Command() ucan.Command {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Cmd
}

// The timestamp at which the invocation becomes invalid.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#expiration
func (inv *Invocation) Expiration() *ucan.UTCUnixTimestamp {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Exp
}

// An issuance timestamp.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#issued-at
func (inv *Invocation) IssuedAt() *ucan.UTCUnixTimestamp {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Iat
}

// Issuer DID (sender).
//
// https://github.com/ucan-wg/spec/blob/main/README.md#issuer--audience
func (inv *Invocation) Issuer() ucan.Principal {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Iss
}

// Arbitrary metadata or extensible fields.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#metadata
func (inv *Invocation) Metadata() ipld.Map[string, any] {
	if inv.meta == nil {
		return nil
	}
	return inv.meta
}

// The datamodel this invocation is built from.
func (inv *Invocation) Model() *idm.EnvelopeModel {
	return inv.model
}

// A unique, random nonce. It ensures that multiple (non-idempotent) invocations
// are unique. The nonce SHOULD be empty (0x) for Commands that are idempotent
// (such as deterministic Wasm modules or standards-abiding HTTP PUT requests).
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#nonce
func (inv *Invocation) Nonce() ucan.Nonce {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Nonce
}

// The path of authority from the subject to the invoker.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#proofs
func (inv *Invocation) Proofs() []cid.Cid {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Prf
}

// The signature over the payload.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#envelope
func (inv *Invocation) Signature() ucan.Signature {
	return inv.sig
}

// The Subject being invoked.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#subject
func (inv *Invocation) Subject() ucan.Principal {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Sub
}

var _ ucan.Invocation = (*Invocation)(nil)

func Encode(inv ucan.Invocation) ([]byte, error) {
	sigHeader, err := varsig.Encode(inv.Signature().Header())
	if err != nil {
		return nil, fmt.Errorf("encoding varsig header: %w", err)
	}

	var aud *did.DID
	if inv.Audience() != nil {
		a := inv.Audience().DID()
		aud = &a
	}

	var args cbg.Deferred
	if cmargs, ok := inv.Arguments().(dagcbor.CBORMarshaler); ok {
		var buf bytes.Buffer
		err := cmargs.MarshalCBOR(&buf)
		if err != nil {
			return nil, fmt.Errorf("marshaling arguments CBOR: %w", err)
		}
		args.Raw = buf.Bytes()
	} else {
		return nil, errors.New("arguments are not CBOR marshaler")
	}

	var meta *cbg.Deferred
	if inv.Metadata() != nil {
		if cmmeta, ok := inv.Metadata().(dagcbor.CBORMarshaler); ok {
			var buf bytes.Buffer
			err := cmmeta.MarshalCBOR(&buf)
			if err != nil {
				return nil, fmt.Errorf("marshaling metadata CBOR: %w", err)
			}
			meta = &cbg.Deferred{Raw: buf.Bytes()}
		} else {
			return nil, errors.New("metadata is not CBOR marshaler")
		}
	}

	model := &idm.EnvelopeModel{
		Signature: inv.Signature().Bytes(),
		SigPayload: idm.SigPayloadModel{
			Header: sigHeader,
			TokenPayload1_0_0_rc1: &idm.TokenPayloadModel1_0_0_rc1{
				Iss:   inv.Issuer().DID(),
				Sub:   inv.Subject().DID(),
				Aud:   aud,
				Cmd:   inv.Command(),
				Args:  args,
				Prf:   inv.Proofs(),
				Meta:  meta,
				Nonce: inv.Nonce(),
				Exp:   inv.Expiration(),
				Iat:   inv.IssuedAt(),
				Cause: inv.Cause(),
			},
		},
	}

	buf := bytes.NewBuffer([]byte{})
	err = model.MarshalCBOR(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decode(data []byte) (*Invocation, error) {
	model := idm.EnvelopeModel{}
	err := model.UnmarshalCBOR(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("unmarshaling invocation envelope CBOR: %w", err)
	}
	if model.SigPayload.TokenPayload1_0_0_rc1 == nil {
		return nil, errors.New("invalid or unsupported invocation token payload")
	}
	header, err := varsig.Decode(model.SigPayload.Header)
	if err != nil {
		return nil, fmt.Errorf("decoding varsig header: %w", err)
	}
	sig := signature.NewSignature(header, model.Signature)
	var args datamodel.Map
	err = args.UnmarshalCBOR(bytes.NewReader(model.SigPayload.TokenPayload1_0_0_rc1.Args.Raw))
	if err != nil {
		return nil, fmt.Errorf("unmarshaling arguments CBOR: %w", err)
	}
	var meta *datamodel.Map
	if model.SigPayload.TokenPayload1_0_0_rc1.Meta != nil {
		meta = &datamodel.Map{}
		err = meta.UnmarshalCBOR(bytes.NewReader(model.SigPayload.TokenPayload1_0_0_rc1.Meta.Raw))
		if err != nil {
			return nil, fmt.Errorf("unmarshaling metadata CBOR: %w", err)
		}
	}
	return &Invocation{
		sig:   sig,
		model: &model,
		args:  &args,
		meta:  meta,
	}, nil
}

func Invoke(
	issuer ucan.Signer,
	subject ucan.Subject,
	command ucan.Command,
	arguments ipld.Map[string, any],
	options ...Option,
) (*Invocation, error) {
	cfg := invocationConfig{}
	for _, opt := range options {
		opt(&cfg)
	}

	if issuer.SignatureCode() != ed25519.Code {
		return nil, fmt.Errorf("unknown signature code: %d", issuer.SignatureCode())
	}
	h, err := varsig.Encode(common.Ed25519DagCbor)
	if err != nil {
		return nil, fmt.Errorf("encoding varsig header: %w", err)
	}

	cmd, err := cmd.Parse(string(command))
	if err != nil {
		return nil, fmt.Errorf("parsing command: %w", err)
	}

	var args cbg.Deferred
	if cmargs, ok := arguments.(dagcbor.CBORMarshaler); ok {
		var buf bytes.Buffer
		err := cmargs.MarshalCBOR(&buf)
		if err != nil {
			return nil, fmt.Errorf("marshaling arguments CBOR: %w", err)
		}
		args.Raw = buf.Bytes()
	} else {
		return nil, errors.New("arguments are not CBOR marshaler")
	}

	var meta *cbg.Deferred
	if cfg.meta != nil {
		if cmmeta, ok := cfg.meta.(dagcbor.CBORMarshaler); ok {
			var buf bytes.Buffer
			err := cmmeta.MarshalCBOR(&buf)
			if err != nil {
				return nil, fmt.Errorf("marshaling metadata CBOR: %w", err)
			}
			meta = &cbg.Deferred{Raw: buf.Bytes()}
		} else {
			return nil, errors.New("metadata is not CBOR marshaler")
		}
	}

	nnc := cfg.nnc
	if nnc == nil {
		if cfg.nonnc {
			nnc = []byte{}
		} else {
			nnc = nonce.Generate(16)
		}
	}

	var exp *ucan.UTCUnixTimestamp
	if !cfg.noexp {
		if cfg.exp == nil {
			in30s := uint64(ucan.Now() + 30)
			exp = &in30s
		} else {
			exp = cfg.exp
		}
	}

	iat := ucan.Now()

	tokenPayload := &idm.TokenPayloadModel1_0_0_rc1{
		Iss:   issuer.DID(),
		Sub:   subject.DID(),
		Aud:   cfg.aud,
		Cmd:   cmd,
		Args:  args,
		Prf:   cfg.prf,
		Meta:  meta,
		Nonce: nnc,
		Exp:   exp,
		Iat:   &iat,
		Cause: cfg.cause,
	}

	sigPayload := idm.SigPayloadModel{
		Header:                h,
		TokenPayload1_0_0_rc1: tokenPayload,
	}

	var buf bytes.Buffer
	err = sigPayload.MarshalCBOR(&buf)
	if err != nil {
		return nil, fmt.Errorf("marshaling token payload: %w", err)
	}

	sigBytes := issuer.Sign(buf.Bytes())
	sig := signature.NewSignature(common.Ed25519DagCbor, sigBytes)

	model := idm.EnvelopeModel{
		Signature:  sigBytes,
		SigPayload: sigPayload,
	}

	var argsMap *datamodel.Map
	if a, ok := arguments.(*datamodel.Map); ok {
		argsMap = a
	} else {
		err := argsMap.UnmarshalCBOR(bytes.NewReader(args.Raw))
		if err != nil {
			return nil, err
		}
	}

	var metaMap *datamodel.Map
	if cfg.meta != nil {
		if m, ok := cfg.meta.(*datamodel.Map); ok {
			metaMap = m
		} else {
			err := metaMap.UnmarshalCBOR(bytes.NewReader(meta.Raw))
			if err != nil {
				return nil, err
			}
		}
	}

	return &Invocation{
		sig:   sig,
		model: &model,
		args:  argsMap,
		meta:  metaMap,
	}, nil
}
