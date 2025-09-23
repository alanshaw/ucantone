package datamodel

import (
	"fmt"
	"io"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ucan"
	edm "github.com/alanshaw/ucantone/ucan/envelope/datamodel"
	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
)

// CBORMarshalable is an interface describing a type that allows both
// marshalling to CBOR as well as unmarshalling from CBOR.
type CBORMarshalable interface {
	cbg.CBORMarshaler
	cbg.CBORUnmarshaler
}

type StructModel struct {
	Value CBORMarshalable
}

func (sm *StructModel) MarshalCBOR(w io.Writer) error {
	if sm.Value == nil {
		_, err := w.Write(cbg.CborEncodeMajorType(cbg.MajMap, 0))
		return err
	}
	return sm.Value.MarshalCBOR(w)
}

func (sm *StructModel) UnmarshalCBOR(r io.Reader) error {
	if sm.Value == nil {
		*sm = StructModel{}
		cr := cbg.NewCborReader(r)
		maj, extra, err := cr.ReadHeader()
		if err != nil {
			return err
		}
		defer func() {
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
		}()
		if maj != cbg.MajMap {
			return fmt.Errorf("cbor input should be of type map")
		}
		if extra != 0 {
			return fmt.Errorf("StructModel: map struct too large (%d)", extra)
		}
		return nil
	}
	return sm.Value.UnmarshalCBOR(r)
}

var _ CBORMarshalable = (*StructModel)(nil)

type ArgsModel = StructModel

type MetaModel = StructModel

type TokenPayloadModel1_0_0_rc1 struct {
	// Issuer DID (sender).
	Iss did.DID `cborgen:"iss"`
	// The Subject being invoked.
	Sub ucan.Subject `cborgen:"sub"`
	// The DID of the intended Executor if different from the Subject.
	Aud *did.DID `cborgen:"aud,omitempty"`
	// The command to invoke.
	Cmd ucan.Command `cborgen:"cmd"`
	// The command arguments.
	Args ArgsModel `cborgen:"args"`
	// Delegations that prove the chain of authority.
	Prf []cid.Cid `cborgen:"prf"`
	// Arbitrary metadata.
	Meta *MetaModel `cborgen:"meta,omitempty"`
	// A unique, random nonce.
	Nonce ucan.Nonce `cborgen:"nonce"`
	// The timestamp at which the invocation becomes invalid.
	Exp *ucan.UTCUnixTimestamp `cborgen:"exp"`
	// The timestamp at which the invocation was created.
	Iat *ucan.UTCUnixTimestamp `cborgen:"iat,omitempty"`
	// CID of the receipt that enqueued the Task.
	Cause *cid.Cid `cborgen:"cause,omitempty"`
}

type SigPayloadModel struct {
	// The Varsig v1 header.
	Header []byte `cborgen:"h"`
	// The UCAN token payload.
	TokenPayload1_0_0_rc1 *TokenPayloadModel1_0_0_rc1 `cborgen:"ucan/inv@1.0.0-rc.1,omitempty"`
}

type EnvelopeModel edm.EnvelopeModel[SigPayloadModel]
