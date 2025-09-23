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
	b, err := Encode(d)
	if err != nil {
		return err
	}
	return cbg.WriteByteArray(w, b)
}

func (d *DID) UnmarshalCBOR(r io.Reader) error {
	b, err := cbg.ReadByteArray(r, 2048)
	if err != nil {
		return err
	}
	decoded, err := Decode(b)
	if err != nil {
		return err
	}
	*d = decoded
	return nil
}

func Encode(d DID) ([]byte, error) {
	str := d.str
	if !strings.HasPrefix(str, Prefix) {
		fmt.Println("Encode", d.str)
		return nil, fmt.Errorf("must start with 'did:'")
	}

	if strings.HasPrefix(str, KeyPrefix) {
		code, bytes, err := mbase.Decode(str[len(KeyPrefix):])
		if err != nil {
			return nil, err
		}
		if code != mbase.Base58BTC {
			return nil, fmt.Errorf("not Base58BTC encoded")
		}
		return bytes, nil
	}

	buf := make([]byte, MethodOffset)
	varint.PutUvarint(buf, DIDCore)
	suffix, _ := strings.CutPrefix(str, Prefix)
	buf = append(buf, suffix...)
	return buf, nil
}

func Decode(bytes []byte) (DID, error) {
	code, n, err := varint.FromUvarint(bytes)
	if err != nil {
		return DID{}, err
	}
	switch code {
	case Ed25519, RSA:
		b58key, _ := mbase.Encode(mbase.Base58BTC, bytes)
		return DID{KeyPrefix + b58key}, nil
	case DIDCore:
		return DID{Prefix + string(bytes[n:])}, nil
	}
	return DID{}, fmt.Errorf("unsupported DID encoding: 0x%x", code)
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
