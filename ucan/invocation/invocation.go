package invocation

import (
	"bytes"
	"fmt"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
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
var NoArguments = &idm.NoArgumentsModel{}

type Invocation[Args, Meta dagcbor.CBORMarshalable] struct {
	sig   *signature.Signature
	model *idm.EnvelopeModel
}

// Parameters expected by the command.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#arguments
func (inv *Invocation[Args, Meta]) Arguments() Args {
	var args Args
	err := args.UnmarshalCBOR(bytes.NewReader(inv.model.SigPayload.TokenPayload1_0_0_rc1.Args.Raw))
	if err != nil {
		fmt.Println(err)
	}
	return args
}

// The DID of the intended Executor if different from the Subject. May be nil.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#issuer--audience
func (inv *Invocation[Args, Meta]) Audience() ucan.Principal {
	aud := inv.model.SigPayload.TokenPayload1_0_0_rc1.Aud
	if aud == nil {
		return nil
	}
	return aud
}

// A provenance claim describing which receipt requested it.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#cause
func (inv *Invocation[Args, Meta]) Cause() *cid.Cid {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Cause
}

// The command to invoke.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#command
func (inv *Invocation[Args, Meta]) Command() ucan.Command {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Cmd
}

// The timestamp at which the invocation becomes invalid.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#expiration
func (inv *Invocation[Args, Meta]) Expiration() *ucan.UTCUnixTimestamp {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Exp
}

// An issuance timestamp.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#issued-at
func (inv *Invocation[Args, Meta]) IssuedAt() *ucan.UTCUnixTimestamp {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Iat
}

// Issuer DID (sender).
//
// https://github.com/ucan-wg/spec/blob/main/README.md#issuer--audience
func (inv *Invocation[Args, Meta]) Issuer() ucan.Principal {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Iss
}

// Arbitrary metadata or extensible fields.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#metadata
func (inv *Invocation[Args, Meta]) Metadata() Meta {
	var meta Meta
	err := meta.UnmarshalCBOR(bytes.NewReader(inv.model.SigPayload.TokenPayload1_0_0_rc1.Meta.Raw))
	if err != nil {
		fmt.Println(err)
	}
	return meta
}

// The datamodel this invocation is built from.
func (inv *Invocation[Args, Meta]) Model() *idm.EnvelopeModel {
	return inv.model
}

// A unique, random nonce. It ensures that multiple (non-idempotent) invocations
// are unique. The nonce SHOULD be empty (0x) for Commands that are idempotent
// (such as deterministic Wasm modules or standards-abiding HTTP PUT requests).
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#nonce
func (inv *Invocation[Args, Meta]) Nonce() ucan.Nonce {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Nonce
}

// The path of authority from the subject to the invoker.
//
// https://github.com/ucan-wg/invocation/blob/main/README.md#proofs
func (inv *Invocation[Args, Meta]) Proofs() []cid.Cid {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Prf
}

// The signature over the payload.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#envelope
func (inv *Invocation[Args, Meta]) Signature() ucan.Signature {
	return inv.sig
}

// The Subject being invoked.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#subject
func (inv *Invocation[Args, Meta]) Subject() ucan.Principal {
	return inv.model.SigPayload.TokenPayload1_0_0_rc1.Sub
}

var _ ucan.Invocation[dagcbor.CBORMarshalable, dagcbor.CBORMarshalable] = (*Invocation[dagcbor.CBORMarshalable, dagcbor.CBORMarshalable])(nil)

func Encode[Args, Meta dagcbor.CBORMarshalable](inv ucan.Invocation[Args, Meta]) ([]byte, error) {
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

	var argsBuf bytes.Buffer
	err = inv.Arguments().MarshalCBOR(&argsBuf)
	if err != nil {
		return nil, fmt.Errorf("marshaling CBOR: %w", err)
	}
	args.Raw = argsBuf.Bytes()

	var meta *cbg.Deferred
	if inv.Metadata() != nil {
		m := inv.Metadata()
		var buf bytes.Buffer
		err := m.MarshalCBOR(&buf)
		if err != nil {
			return nil, fmt.Errorf("marshaling CBOR: %w", err)
		}
		meta = &cbg.Deferred{Raw: buf.Bytes()}
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

func Decode[Args, Meta dagcbor.CBORMarshalable](data []byte) (*Invocation[Args, Meta], error) {
	model := idm.EnvelopeModel{}
	err := model.UnmarshalCBOR(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("unmarshaling CBOR: %w", err)
	}
	header, err := varsig.Decode(model.SigPayload.Header)
	if err != nil {
		return nil, fmt.Errorf("decoding varsig header: %w", err)
	}
	sig := signature.NewSignature(header, model.Signature)
	return &Invocation[Args, Meta]{sig: sig, model: &model}, nil
}

func Invoke[Args, Meta dagcbor.CBORMarshalable](
	issuer ucan.Signer,
	subject ucan.Subject,
	command ucan.Command,
	arguments Args,
	options ...Option,
) (*Invocation[Args, Meta], error) {
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

	var argsBuf bytes.Buffer
	err = arguments.MarshalCBOR(&argsBuf)
	if err != nil {
		return nil, fmt.Errorf("marshaling arguments CBOR: %w", err)
	}
	args := cbg.Deferred{Raw: argsBuf.Bytes()}

	var meta *cbg.Deferred
	if cfg.meta != nil {
		var buf bytes.Buffer
		err = arguments.MarshalCBOR(&buf)
		if err != nil {
			return nil, fmt.Errorf("marshaling metadata CBOR: %w", err)
		}
		meta = &cbg.Deferred{Raw: argsBuf.Bytes()}
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
		Sub:   subject,
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
	err = tokenPayload.MarshalCBOR(&buf)
	if err != nil {
		return nil, fmt.Errorf("marshaling token payload: %w", err)
	}

	sigBytes := issuer.Sign(buf.Bytes())
	sig := signature.NewSignature(common.Ed25519DagCbor, sigBytes)

	model := idm.EnvelopeModel{
		Signature:  sigBytes,
		SigPayload: sigPayload,
	}

	return &Invocation[Args, Meta]{sig: sig, model: &model}, nil
}
