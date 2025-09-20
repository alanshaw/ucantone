package did

import (
	"encoding/json"
	"fmt"
	"strings"

	mbase "github.com/multiformats/go-multibase"
	varint "github.com/multiformats/go-varint"
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
type DID string

func (d DID) DID() DID {
	return d
}

// String formats the decentralized identity document (DID) as a string.
func (d DID) String() string {
	return string(d)
}

func (d DID) MarshalJSON() ([]byte, error) {
	if d == "" {
		return json.Marshal(nil)
	}
	return json.Marshal(d.String())
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

func Encode(d DID) ([]byte, error) {
	str := string(d)
	if !strings.HasPrefix(str, Prefix) {
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
		return "", err
	}
	switch code {
	case Ed25519, RSA:
		b58key, _ := mbase.Encode(mbase.Base58BTC, bytes)
		return DID(KeyPrefix + b58key), nil
	case DIDCore:
		return DID(Prefix + string(bytes[n:])), nil
	}
	return "", fmt.Errorf("unsupported DID encoding: 0x%x", code)
}

func Parse(str string) (DID, error) {
	if !strings.HasPrefix(str, Prefix) {
		return "", fmt.Errorf("must start with 'did:'")
	}
	if strings.HasPrefix(str, KeyPrefix) {
		code, _, err := mbase.Decode(str[len(KeyPrefix):])
		if err != nil {
			return "", err
		}
		if code != mbase.Base58BTC {
			return "", fmt.Errorf("not Base58BTC encoded")
		}
	}
	return DID(str), nil
}
