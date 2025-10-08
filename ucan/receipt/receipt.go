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
	"github.com/alanshaw/ucantone/ucan/invocation"
	rdm "github.com/alanshaw/ucantone/ucan/receipt/datamodel"
	cid "github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
)

const Command = command.Command("/ucan/assert/receipt")

type Receipt struct {
	invocation.Invocation
	ran cid.Cid
	out result.Result[ipld.Any, ipld.Any]
}

// Out is the attested result of the execution of the task.
func (rcpt *Receipt) Out() result.Result[ipld.Any, ipld.Any] {
	return rcpt.out
}

// Ran is the CID of the executed task this receipt is for.
func (rcpt *Receipt) Ran() cid.Cid {
	return rcpt.ran
}

var _ ucan.Receipt = (*Receipt)(nil)

func Encode(rcpt ucan.Receipt) ([]byte, error) {
	return rcpt.Bytes(), nil
}

func Decode(data []byte) (*Receipt, error) {
	inv, err := invocation.Decode(data)
	if err != nil {
		return nil, err
	}

	invArgs, ok := inv.Arguments().(dagcbor.CBORMarshaler)
	if !ok {
		return nil, errors.New("invocation arguments are not CBOR unmarahsler")
	}

	if inv.Command() != Command {
		return nil, fmt.Errorf("invalid receipt command %s, expected %s", inv.Command().String(), Command.String())
	}

	var receiptArgs rdm.ArgsModel
	err = datamodel.Rebind(invArgs, &receiptArgs)
	if err != nil {
		return nil, fmt.Errorf("decoding receipt arguments: %w", err)
	}

	var out result.Result[any, any]
	if receiptArgs.Out.Ok != nil {
		var a datamodel.Any
		err := a.UnmarshalCBOR(bytes.NewReader(receiptArgs.Out.Ok.Raw))
		if err != nil {
			return nil, fmt.Errorf("unmarshaling ok result CBOR: %w", err)
		}
		out = result.Ok[any, any](a.Value)
	} else if receiptArgs.Out.Err != nil {
		var a datamodel.Any
		err := a.UnmarshalCBOR(bytes.NewReader(receiptArgs.Out.Err.Raw))
		if err != nil {
			return nil, fmt.Errorf("unmarshaling error result CBOR: %w", err)
		}
		out = result.Error[any](a.Value)
	} else {
		return nil, errors.New("invalid result, neither ok nor error")
	}

	return &Receipt{Invocation: *inv, ran: receiptArgs.Ran, out: out}, nil
}

// Issue creates a new receipt: an attestation that a task was run and it
// resulted in the passed output.
func Issue[O, X ipld.Any](
	executor ucan.Signer,
	ran cid.Cid,
	out result.Result[O, X],
	options ...Option,
) (*Receipt, error) {
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

	var args datamodel.Map
	err = datamodel.Rebind(&rdm.ArgsModel{
		Ran: ran,
		Out: outModel,
	}, &args)
	if err != nil {
		return nil, err
	}

	options = append(options, invocation.WithAudience(executor))

	inv, err := invocation.Invoke(executor, executor.DID(), Command, &args, options...)
	if err != nil {
		return nil, err
	}

	return &Receipt{
		Invocation: *inv,
		ran:        ran,
		out: result.MapResultR0(
			out,
			func(o O) any { return o },
			func(x X) any { return x },
		),
	}, nil
}
