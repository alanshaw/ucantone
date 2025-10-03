package receipt

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/result"
	rsdm "github.com/alanshaw/ucantone/result/datamodel"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/crypto/signature"
	"github.com/alanshaw/ucantone/ucan/nonce"
	rdm "github.com/alanshaw/ucantone/ucan/receipt/datamodel"
	"github.com/alanshaw/ucantone/varsig"
	"github.com/alanshaw/ucantone/varsig/algorithm/ed25519"
	"github.com/alanshaw/ucantone/varsig/common"
	cid "github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
)

const Command = command.Command("/ucan/assert/receipt")

type Receipt struct {
	sig   *signature.Signature
	model *rdm.EnvelopeModel
	out   result.Result[any, any]
	meta  *datamodel.Map
}

// Ran is the CID of the executed task this receipt is for.
func (rcpt *Receipt) Ran() cid.Cid {
	return rcpt.model.SigPayload.TokenPayload1_0_0_rc1.Args.Ran
}

// Out is the attested result of the execution of the task.
func (rcpt *Receipt) Out() result.Result[any, any] {
	return rcpt.out
}

func (rcpt *Receipt) Audience() ucan.Principal {
	return rcpt.model.SigPayload.TokenPayload1_0_0_rc1.Aud
}

func (rcpt *Receipt) Command() ucan.Command {
	return rcpt.model.SigPayload.TokenPayload1_0_0_rc1.Cmd
}

func (rcpt *Receipt) Expiration() *ucan.UTCUnixTimestamp {
	return rcpt.model.SigPayload.TokenPayload1_0_0_rc1.Exp
}

func (rcpt *Receipt) IssuedAt() *ucan.UTCUnixTimestamp {
	return rcpt.model.SigPayload.TokenPayload1_0_0_rc1.Iat
}

func (rcpt *Receipt) Issuer() ucan.Principal {
	return rcpt.model.SigPayload.TokenPayload1_0_0_rc1.Iss
}

func (rcpt *Receipt) Metadata() ipld.Map[string, any] {
	if rcpt.meta == nil {
		return nil
	}
	return rcpt.meta
}

// The datamodel this receipt is built from.
func (rcpt *Receipt) Model() *rdm.EnvelopeModel {
	return rcpt.model
}

func (rcpt *Receipt) Nonce() ucan.Nonce {
	return rcpt.model.SigPayload.TokenPayload1_0_0_rc1.Nonce
}

func (rcpt *Receipt) Proofs() []cid.Cid {
	return rcpt.model.SigPayload.TokenPayload1_0_0_rc1.Prf
}

func (rcpt *Receipt) Signature() ucan.Signature {
	return rcpt.sig
}

func (rcpt *Receipt) Subject() ucan.Principal {
	return rcpt.model.SigPayload.TokenPayload1_0_0_rc1.Sub
}

var _ ucan.Receipt = (*Receipt)(nil)

func Encode(rcpt ucan.Receipt) ([]byte, error) {
	sigHeader, err := varsig.Encode(rcpt.Signature().Header())
	if err != nil {
		return nil, fmt.Errorf("encoding varsig header: %w", err)
	}

	out, err := result.MatchResultR2(
		rcpt.Out(),
		func(o any) (rsdm.ResultModel, error) {
			var b bytes.Buffer
			err := datamodel.NewAny(o).MarshalCBOR(&b)
			if err != nil {
				return rsdm.ResultModel{}, fmt.Errorf("marshaling result ok value: %w", err)
			}
			return rsdm.ResultModel{Ok: &cbg.Deferred{Raw: b.Bytes()}}, nil
		},
		func(x any) (rsdm.ResultModel, error) {
			var b bytes.Buffer
			err := datamodel.NewAny(x).MarshalCBOR(&b)
			if err != nil {
				return rsdm.ResultModel{}, fmt.Errorf("marshaling result error value: %w", err)
			}
			return rsdm.ResultModel{Err: &cbg.Deferred{Raw: b.Bytes()}}, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("encoding result: %w", err)
	}

	args := rdm.ArgsModel{
		Ran: rcpt.Ran(),
		Out: out,
	}

	var meta *cbg.Deferred
	if rcpt.Metadata() != nil {
		if cmmeta, ok := rcpt.Metadata().(dagcbor.CBORMarshaler); ok {
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

	model := &rdm.EnvelopeModel{
		Signature: rcpt.Signature().Bytes(),
		SigPayload: rdm.SigPayloadModel{
			Header: sigHeader,
			TokenPayload1_0_0_rc1: &rdm.TokenPayloadModel1_0_0_rc1{
				Iss:   rcpt.Issuer().DID(),
				Sub:   rcpt.Subject().DID(),
				Aud:   rcpt.Audience().DID(),
				Cmd:   rcpt.Command(),
				Args:  args,
				Prf:   rcpt.Proofs(),
				Meta:  meta,
				Nonce: rcpt.Nonce(),
				Exp:   rcpt.Expiration(),
				Iat:   rcpt.IssuedAt(),
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

func Decode(data []byte) (*Receipt, error) {
	model := rdm.EnvelopeModel{}
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

	var out result.Result[any, any]
	if model.SigPayload.TokenPayload1_0_0_rc1.Args.Out.Ok != nil {
		var a datamodel.Any
		err := a.UnmarshalCBOR(bytes.NewReader(model.SigPayload.TokenPayload1_0_0_rc1.Args.Out.Ok.Raw))
		if err != nil {
			return nil, fmt.Errorf("unmarshaling ok result CBOR: %w", err)
		}
		out = result.Ok[any, any](a.Value)
	} else if model.SigPayload.TokenPayload1_0_0_rc1.Args.Out.Err != nil {
		var a datamodel.Any
		err := a.UnmarshalCBOR(bytes.NewReader(model.SigPayload.TokenPayload1_0_0_rc1.Args.Out.Err.Raw))
		if err != nil {
			return nil, fmt.Errorf("unmarshaling error result CBOR: %w", err)
		}
		out = result.Error[any](a.Value)
	} else {
		return nil, errors.New("invalid result, neither ok nor error")
	}

	var meta *datamodel.Map
	if model.SigPayload.TokenPayload1_0_0_rc1.Meta != nil {
		meta = &datamodel.Map{}
		err = meta.UnmarshalCBOR(bytes.NewReader(model.SigPayload.TokenPayload1_0_0_rc1.Meta.Raw))
		if err != nil {
			return nil, fmt.Errorf("unmarshaling metadata CBOR: %w", err)
		}
	}

	return &Receipt{
		sig:   sig,
		model: &model,
		out:   out,
		meta:  meta,
	}, nil
}

func Issue[O, X any](
	executor ucan.Signer,
	ran cid.Cid,
	out result.Result[O, X],
	options ...Option,
) (*Receipt, error) {
	cfg := receiptConfig{}
	for _, opt := range options {
		opt(&cfg)
	}

	if executor.SignatureCode() != ed25519.Code {
		return nil, fmt.Errorf("unknown signature code: %d", executor.SignatureCode())
	}
	h, err := varsig.Encode(common.Ed25519DagCbor)
	if err != nil {
		return nil, fmt.Errorf("encoding varsig header: %w", err)
	}

	outModel, err := result.MatchResultR2(
		out,
		func(o O) (rsdm.ResultModel, error) {
			var b bytes.Buffer
			err := datamodel.NewAny(o).MarshalCBOR(&b)
			if err != nil {
				return rsdm.ResultModel{}, fmt.Errorf("marshaling result ok value: %w", err)
			}
			return rsdm.ResultModel{Ok: &cbg.Deferred{Raw: b.Bytes()}}, nil
		},
		func(x X) (rsdm.ResultModel, error) {
			var b bytes.Buffer
			err := datamodel.NewAny(x).MarshalCBOR(&b)
			if err != nil {
				return rsdm.ResultModel{}, fmt.Errorf("marshaling result error value: %w", err)
			}
			return rsdm.ResultModel{Err: &cbg.Deferred{Raw: b.Bytes()}}, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("encoding result: %w", err)
	}

	args := rdm.ArgsModel{
		Ran: ran,
		Out: outModel,
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

	tokenPayload := &rdm.TokenPayloadModel1_0_0_rc1{
		Iss:   executor.DID(),
		Sub:   executor.DID(),
		Aud:   executor.DID(),
		Cmd:   Command,
		Args:  args,
		Prf:   cfg.prf,
		Meta:  meta,
		Nonce: nnc,
		Exp:   exp,
		Iat:   &iat,
	}

	sigPayload := rdm.SigPayloadModel{
		Header:                h,
		TokenPayload1_0_0_rc1: tokenPayload,
	}

	var buf bytes.Buffer
	err = sigPayload.MarshalCBOR(&buf)
	if err != nil {
		return nil, fmt.Errorf("marshaling token payload: %w", err)
	}

	sigBytes := executor.Sign(buf.Bytes())
	sig := signature.NewSignature(common.Ed25519DagCbor, sigBytes)

	model := rdm.EnvelopeModel{
		Signature:  sigBytes,
		SigPayload: sigPayload,
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

	return &Receipt{
		sig:   sig,
		model: &model,
		out: result.MapResultR0(
			out,
			func(o O) any { return o },
			func(x X) any { return x },
		),
		meta: metaMap,
	}, nil
}
