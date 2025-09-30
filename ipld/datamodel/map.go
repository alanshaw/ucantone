package datamodel

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"slices"
	"sort"
	"strings"

	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
)

// Map is a CBOR backed implementation of [ipld.Map]. Keys are strings and
// values may be any of the types supported by [Any].
type Map struct {
	keys   []string
	values map[string]cbg.Deferred
}

type MapOption func(m *Map) error

// WithValue adds the passed value to the new map. The value may be any of the
// types supported by [Any].
func WithValue(key string, value any) MapOption {
	return func(m *Map) error {
		return m.SetValue(key, value)
	}
}

// WithValues adds the passed values to the new map. The values may be any of
// the types supported by [Any].
func WithValues(values map[string]any) MapOption {
	return func(m *Map) error {
		for k, v := range values {
			err := m.SetValue(k, v)
			if err != nil {
				return fmt.Errorf("setting value for key %s: %w", k, err)
			}
		}
		return nil
	}
}

func NewMap(options ...MapOption) (*Map, error) {
	m := Map{values: map[string]cbg.Deferred{}}
	for _, opt := range options {
		err := opt(&m)
		if err != nil {

		}
	}
	return &m, nil
}

// NewMapFromCBORMarshaler creates a new [ipld.Map] from the passed CBOR
// marshaler object. The object MUST marshal to a CBOR map type. It's values may
// be any of the types supported by [Any].
func NewMapFromCBORMarshaler(data dagcbor.CBORMarshaler) (*Map, error) {
	var buf bytes.Buffer
	err := data.MarshalCBOR(&buf)
	if err != nil {
		return nil, fmt.Errorf("marshalling CBOR: %w", err)
	}
	var m Map
	err = m.UnmarshalCBOR(&buf)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling CBOR: %w", err)
	}
	return &m, nil
}

func (m *Map) Keys() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, k := range m.keys {
			if !yield(k) {
				return
			}
		}
	}
}

func (m *Map) Value(k string) (any, bool) {
	v, ok := m.values[k]
	if !ok {
		return nil, false
	}
	a := &Any{}
	a.UnmarshalCBOR(bytes.NewReader(v.Raw))
	return a.Value, true
}

func (m *Map) SetValue(k string, v any) error {
	a := Any{Value: v}
	var buf bytes.Buffer
	if err := a.MarshalCBOR(&buf); err != nil {
		return err
	}
	_, ok := m.values[k]
	m.values[k] = cbg.Deferred{Raw: buf.Bytes()}
	if !ok {
		m.keys = append(m.keys, k)
	}
	return nil
}

func (m *Map) MarshalCBOR(w io.Writer) error {
	if m == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)

	if err := cw.WriteMajorTypeHeader(cbg.MajMap, uint64(len(m.keys))); err != nil {
		return err
	}

	keys := make([]string, len(m.keys))
	copy(keys, m.keys)
	sort.Slice(keys, func(i, j int) bool {
		fi := keys[i]
		fj := keys[j]
		if len(fi) < len(fj) {
			return true
		}
		if len(fi) > len(fj) {
			return false
		}
		return fi < fj
	})

	for _, k := range keys {
		if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len(k))); err != nil {
			return err
		}
		if _, err := cw.WriteString(k); err != nil {
			return err
		}

		v := m.values[k]
		if err := v.MarshalCBOR(w); err != nil {
			return fmt.Errorf("marshalling map value for key: %s: %w", k, err)
		}
	}

	return nil
}

func (m *Map) UnmarshalCBOR(r io.Reader) (err error) {
	*m = Map{values: map[string]cbg.Deferred{}}

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

	if extra > cbg.MaxLength {
		return fmt.Errorf("Map: map struct too large (%d)", extra)
	}

	n := extra
	nameBuf := make([]byte, 2048)
	for range n {
		nameLen, ok, err := cbg.ReadFullStringIntoBuf(cr, nameBuf, 8192)
		if err != nil {
			return err
		}

		if !ok {
			if err := cbg.ScanForLinks(cr, func(cid.Cid) {}); err != nil {
				return err
			}
			continue
		}

		name := string(nameBuf[:nameLen])
		m.keys = append(m.keys, name)

		d := cbg.Deferred{}
		if err := d.UnmarshalCBOR(cr); err != nil {
			return fmt.Errorf("failed to read deferred field %s: %w", name, err)
		}
		m.values[name] = d
	}

	return nil
}

func (m *Map) MarshalJSON() ([]byte, error) {
	var b strings.Builder
	_, err := b.WriteString("{")
	if err != nil {
		return nil, err
	}
	keys := make([]string, len(m.keys))
	copy(keys, m.keys)
	slices.Sort(keys)
	for i, k := range keys {
		kBytes, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		_, err = b.WriteString(fmt.Sprintf("%s:", string(kBytes)))
		if err != nil {
			return nil, err
		}

		a := &Any{}
		err = a.UnmarshalCBOR(bytes.NewReader(m.values[k].Raw))
		if err != nil {
			return nil, err
		}

		switch v := a.Value.(type) {
		case []byte:
			_, err = b.WriteString(formatDAGJSONBytes(v))
			if err != nil {
				return nil, err
			}
		default:
			vBytes, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			_, err = b.Write(vBytes)
			if err != nil {
				return nil, err
			}
		}

		if i < len(keys)-1 {
			_, err = b.WriteString(",")
			if err != nil {
				return nil, err
			}
		}
	}
	_, err = b.WriteString("}")
	if err != nil {
		return nil, err
	}
	return []byte(b.String()), nil
}

func formatDAGJSONBytes(bytes []byte) string {
	return fmt.Sprintf(`{"/":{"bytes":"%s"}}`, base64.StdEncoding.EncodeToString(bytes))
}

var _ ipld.Map[string, any] = (*Map)(nil)
