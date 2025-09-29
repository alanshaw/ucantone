package did

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	mbase "github.com/multiformats/go-multibase"
	varint "github.com/multiformats/go-varint"
	cbg "github.com/whyrusleeping/cbor-gen"
)

const Prefix = "did:"
const KeyPrefix = Prefix + "key:"

const DIDCore = 0x0d1d
const Ed25519 = 0xed
const RSA = 0x1205

var MethodOffset = varint.UvarintSize(uint64(DIDCore))

// DID is a decentralized identity, it has the format:
//
//	"did:%s:%s"
//
// The underlying type is string, so DIDs are safe to compare with == and to use
// as keys in maps.
//
// Note: this is not `type DID string` because cbor-gen does not recognise
// MarshalCBOR or UnmarshalCBOR when type is not struct.
type DID struct {
	str string
}

func (d DID) DID() DID {
	return d
}

// String formats the decentralized identity document (DID) as a string.
func (d DID) String() string {
	return d.str
}

func (d DID) MarshalJSON() ([]byte, error) {
	if d.str == "" {
		return json.Marshal(nil)
	}
	return json.Marshal(d.str)
}

func (d *DID) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return fmt.Errorf("parsing string: %w", err)
	}
	if str == "" {
		return nil
	}
	parsed, err := Parse(str)
	if err != nil {
		return fmt.Errorf("parsing DID: %w", err)
	}
	*d = parsed
	return nil
}

func (d DID) MarshalCBOR(w io.Writer) error {
	if d.str == "" {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	cw := cbg.NewCborWriter(w)
	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len(d.str))); err != nil {
		return err
	}
	_, err := cw.WriteString(d.str)
	return err
}

func (d *DID) UnmarshalCBOR(r io.Reader) error {
	cr := cbg.NewCborReader(r)
	b, err := cr.ReadByte()
	if err != nil {
		return err
	}
	if b != cbg.CborNull[0] {
		if err := cr.UnreadByte(); err != nil {
			return err
		}
		str, err := cbg.ReadStringWithMax(cr, 2048)
		if err != nil {
			return err
		}
		parsed, err := Parse(str)
		if err != nil {
			return err
		}
		*d = parsed
	}
	return nil
}

func Parse(str string) (DID, error) {
	if !strings.HasPrefix(str, Prefix) {
		return DID{}, fmt.Errorf("must start with 'did:'")
	}
	if strings.HasPrefix(str, KeyPrefix) {
		code, _, err := mbase.Decode(str[len(KeyPrefix):])
		if err != nil {
			return DID{}, err
		}
		if code != mbase.Base58BTC {
			return DID{}, fmt.Errorf("not Base58BTC encoded")
		}
	}
	return DID{str}, nil
}

func Format(d DID) (string, error) {
	return d.str, nil
}
