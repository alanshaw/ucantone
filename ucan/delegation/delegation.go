package delegation

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
	ddm "github.com/alanshaw/ucantone/ucan/delegation/datamodel"
	"github.com/alanshaw/ucantone/ucan/nonce"
	"github.com/alanshaw/ucantone/varsig"
	"github.com/alanshaw/ucantone/varsig/algorithm/ed25519"
	"github.com/alanshaw/ucantone/varsig/common"
	cid "github.com/ipfs/go-cid"
	multihash "github.com/multiformats/go-multihash/core"
	cbg "github.com/whyrusleeping/cbor-gen"
)

type Delegation struct {
	link  cid.Cid
	bytes []byte
	sig   *signature.Signature
	model *ddm.EnvelopeModel
	meta  *datamodel.Map
}

// Audience can be conceptualized as the receiver of a postal letter.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#issuer--audience
func (d *Delegation) Audience() ucan.Principal {
	return d.model.SigPayload.TokenPayload1_0_0_rc1.Aud
}

// Bytes returns the dag-cbor encoded bytes of this delegation.
func (d *Delegation) Bytes() []byte {
	return d.bytes
}

// Command is a / delimited path describing the set of commands delegated.
//
// https://github.com/ucan-wg/delegation/blob/main/README.md#command
func (d *Delegation) Command() ucan.Command {
	return d.model.SigPayload.TokenPayload1_0_0_rc1.Cmd
}

// Expiration is the time at which the delegation becomes invalid.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#time-bounds
func (d *Delegation) Expiration() *ucan.UTCUnixTimestamp {
	return d.model.SigPayload.TokenPayload1_0_0_rc1.Exp
}

// Issuer can be conceptualized as the sender of a postal letter.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#issuer--audience
func (d *Delegation) Issuer() ucan.Principal {
	return d.model.SigPayload.TokenPayload1_0_0_rc1.Iss
}

// Link returns the IPLD link that corresponds to the encoded bytes of this
// delegation.
func (d *Delegation) Link() cid.Cid {
	return d.link
}

// A map of arbitrary metadata, facts, and proofs of knowledge.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#metadata
func (d *Delegation) Metadata() ipld.Map[string, ipld.Any] {
	if d.meta == nil {
		return nil
	}
	return d.meta
}

// Nonce helps prevent replay attacks and ensures a unique CID per delegation.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#nonce
func (d *Delegation) Nonce() ucan.Nonce {
	return d.model.SigPayload.TokenPayload1_0_0_rc1.Nonce
}

// NotBefore delays the ability to invoke a UCAN.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#time-bounds
func (d *Delegation) NotBefore() *ucan.UTCUnixTimestamp {
	return d.model.SigPayload.TokenPayload1_0_0_rc1.Nbf
}

// Additional constraints on eventual invocation arguments, expressed in the
// UCAN Policy Language.
//
// https://github.com/ucan-wg/delegation/blob/main/README.md#policy
func (d *Delegation) Policy() ucan.Policy {
	return d.model.SigPayload.TokenPayload1_0_0_rc1.Pol
}

// The signature over the payload.
//
// https://github.com/ucan-wg/spec/blob/main/README.md#envelope
func (d *Delegation) Signature() ucan.Signature {
	return d.sig
}

// The Subject that will eventually be invoked.
//
// https://github.com/ucan-wg/delegation/blob/main/README.md#subject
func (d *Delegation) Subject() ucan.Principal {
	sub := d.model.SigPayload.TokenPayload1_0_0_rc1.Sub
	if sub == (did.DID{}) {
		return nil
	}
	return sub
}

var _ ucan.Delegation = (*Delegation)(nil)

func Encode(dlg ucan.Delegation) ([]byte, error) {
	return dlg.Bytes(), nil
}

func Decode(data []byte) (*Delegation, error) {
	model := ddm.EnvelopeModel{}
	err := model.UnmarshalCBOR(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("unmarshaling delegation envelope CBOR: %w", err)
	}
	if model.SigPayload.TokenPayload1_0_0_rc1 == nil {
		return nil, errors.New("invalid or unsupported delegation token payload")
	}
	header, err := varsig.Decode(model.SigPayload.Header)
	if err != nil {
		return nil, fmt.Errorf("decoding varsig header: %w", err)
	}
	sig := signature.NewSignature(header, model.Signature)
	var meta *datamodel.Map
	if model.SigPayload.TokenPayload1_0_0_rc1.Meta != nil {
		meta = &datamodel.Map{}
		err = meta.UnmarshalCBOR(bytes.NewReader(model.SigPayload.TokenPayload1_0_0_rc1.Meta.Raw))
		if err != nil {
			return nil, fmt.Errorf("unmarshaling metadata CBOR: %w", err)
		}
	}
	root, err := cid.V1Builder{
		Codec:  dagcbor.Code,
		MhType: multihash.SHA2_256,
	}.Sum(data)
	if err != nil {
		return nil, fmt.Errorf("hashing delegation bytes: %w", err)
	}
	return &Delegation{
		link:  root,
		bytes: data,
		sig:   sig,
		model: &model,
		meta:  meta,
	}, nil
}

func Delegate(
	issuer ucan.Signer,
	audience ucan.Principal,
	command ucan.Command,
	options ...Option,
) (*Delegation, error) {
	cfg := delegationConfig{}
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

	if cfg.sub == nil && !cfg.powerline {
		return nil, errors.New("one of subject or powerline must be configured")
	}
	if cfg.sub != nil && cfg.powerline {
		return nil, errors.New("subject and powerline cannot both be configured")
	}
	var sub did.DID
	if cfg.sub != nil {
		sub = *cfg.sub
	}

	cmd, err := cmd.Parse(string(command))
	if err != nil {
		return nil, fmt.Errorf("parsing command: %w", err)
	}

	var meta *cbg.Deferred
	var metaMap *datamodel.Map
	if cfg.meta != nil {
		metaMap = datamodel.NewMap(datamodel.WithEntries(cfg.meta.All()))
		var buf bytes.Buffer
		err = metaMap.MarshalCBOR(&buf)
		if err != nil {
			return nil, fmt.Errorf("marshaling metadata CBOR: %w", err)
		}
		meta = &cbg.Deferred{Raw: buf.Bytes()}
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

	tokenPayload := &ddm.TokenPayloadModel1_0_0_rc1{
		Iss:   issuer.DID(),
		Aud:   audience.DID(),
		Sub:   sub,
		Cmd:   cmd,
		Pol:   cfg.pol,
		Nonce: nnc,
		Meta:  meta,
		Nbf:   cfg.nbf,
		Exp:   exp,
	}

	sigPayload := ddm.SigPayloadModel{
		Header:                h,
		TokenPayload1_0_0_rc1: tokenPayload,
	}

	var sigBuf bytes.Buffer
	err = sigPayload.MarshalCBOR(&sigBuf)
	if err != nil {
		return nil, fmt.Errorf("marshaling token payload: %w", err)
	}

	sigBytes := issuer.Sign(sigBuf.Bytes())
	sig := signature.NewSignature(common.Ed25519DagCbor, sigBytes)

	model := ddm.EnvelopeModel{
		Signature:  sigBytes,
		SigPayload: sigPayload,
	}

	var envBuf bytes.Buffer
	err = model.MarshalCBOR(&envBuf)
	if err != nil {
		return nil, fmt.Errorf("marshaling delegation CBOR: %w", err)
	}
	root, err := cid.V1Builder{
		Codec:  dagcbor.Code,
		MhType: multihash.SHA2_256,
	}.Sum(envBuf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("hashing delegation bytes: %w", err)
	}

	return &Delegation{
		link:  root,
		bytes: envBuf.Bytes(),
		sig:   sig,
		model: &model,
		meta:  metaMap,
	}, nil
}

func VerifySignature(dlg ucan.Delegation, verifier ucan.Verifier) (bool, error) {
	var sub did.DID
	if dlg.Subject() != nil {
		sub = dlg.Subject().DID()
	}

	var meta *cbg.Deferred
	var metaMap *datamodel.Map
	if dlg.Metadata() != nil {
		metaMap = datamodel.NewMap(datamodel.WithEntries(dlg.Metadata().All()))
		var buf bytes.Buffer
		err := metaMap.MarshalCBOR(&buf)
		if err != nil {
			return false, fmt.Errorf("marshaling metadata CBOR: %w", err)
		}
		meta = &cbg.Deferred{Raw: buf.Bytes()}
	}

	tokenPayload := &ddm.TokenPayloadModel1_0_0_rc1{
		Iss:   dlg.Issuer().DID(),
		Aud:   dlg.Audience().DID(),
		Sub:   sub,
		Cmd:   dlg.Command(),
		Pol:   dlg.Policy(),
		Nonce: dlg.Nonce(),
		Meta:  meta,
		Nbf:   dlg.NotBefore(),
		Exp:   dlg.Expiration(),
	}

	h, err := varsig.Encode(dlg.Signature().Header())
	if err != nil {
		return false, fmt.Errorf("encoding varsig header: %w", err)
	}

	sigPayload := ddm.SigPayloadModel{
		Header:                h,
		TokenPayload1_0_0_rc1: tokenPayload,
	}

	var sigBuf bytes.Buffer
	err = sigPayload.MarshalCBOR(&sigBuf)
	if err != nil {
		return false, fmt.Errorf("marshaling signature payload: %w", err)
	}

	return dlg.Issuer().DID() == verifier.DID() && verifier.Verify(sigBuf.Bytes(), dlg.Signature().Bytes()), nil
}
