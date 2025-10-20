package datamodel

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"slices"
	"sort"
	"strings"

	"github.com/alanshaw/ucantone/ipld"
	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
)

// Map is a CBOR backed implementation of [ipld.Map]. Keys are strings and
// values may be any of the types supported by [ipld.Any].
type Map struct {
	keys   []string
	values map[string]*Any
}

type MapOption func(m *Map)

// WithEntry adds the passed key/value pair to the new map. The value may be any
// of the types supported by [ipld.Any].
func WithEntry(key string, value ipld.Any) MapOption {
	return func(m *Map) {
		m.Set(key, value)
	}
}

// WithEntries adds the passed key/value pairs to the new map. The values may be
// any of the types supported by [ipld.Any].
func WithEntries(entries iter.Seq2[string, ipld.Any]) MapOption {
	return func(m *Map) {
		for k, v := range entries {
			m.Set(k, v)
		}
	}
}

func NewMap(options ...MapOption) *Map {
	m := Map{values: map[string]*Any{}}
	for _, opt := range options {
		opt(&m)
	}
	return &m
}

func (m *Map) Entries() iter.Seq2[string, ipld.Any] {
	return func(yield func(string, ipld.Any) bool) {
		for _, k := range m.keys {
			v := m.values[k].Value
			if !yield(k, v) {
				return
			}
		}
	}
}

func (m *Map) Get(k string) (ipld.Any, bool) {
	v, ok := m.values[k]
	if !ok {
		return nil, false
	}
	return v.Value, true
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

func (m *Map) Set(k string, v ipld.Any) {
	a := Any{Value: v}
	_, ok := m.values[k]
	m.values[k] = &a
	if !ok {
		m.keys = append(m.keys, k)
	}
}

func (m *Map) Values() iter.Seq[ipld.Any] {
	return func(yield func(ipld.Any) bool) {
		for _, v := range m.values {
			if !yield(v.Value) {
				return
			}
		}
	}
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
			return fmt.Errorf(`marshaling map value for key "%s": %w`, k, err)
		}
	}

	return nil
}

func (m *Map) UnmarshalCBOR(r io.Reader) (err error) {
	*m = Map{values: map[string]*Any{}}

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

		var a Any
		if err := a.UnmarshalCBOR(cr); err != nil {
			return fmt.Errorf(`unmarshaling map value for key "%s": %w`, name, err)
		}
		m.values[name] = &a
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

		switch v := m.values[k].Value.(type) {
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

var _ ipld.Map[string, ipld.Any] = (*Map)(nil)
var _ ipld.MutableMap[string, ipld.Any] = (*Map)(nil)
