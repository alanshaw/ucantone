package datamodel

import (
	"bytes"
	"fmt"
	"io"
	"iter"

	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
)

type Map struct {
	keys   []string
	values map[string]cbg.Deferred
}

func NewMap(data dagcbor.CBORMarshaler) (*Map, error) {
	var buf bytes.Buffer
	err := data.MarshalCBOR(&buf)
	if err != nil {
		return nil, fmt.Errorf("marshalling CBOR: %w", err)
	}
	m := &Map{}
	err = m.UnmarshalCBOR(&buf)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling CBOR: %w", err)
	}
	return m, nil
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

func (m *Map) Value(k string) any {
	v := m.values[k]
	a := &Any{}
	a.UnmarshalCBOR(bytes.NewReader(v.Raw))
	return a.Value
}

func (m *Map) MarshalCBOR(w io.Writer) error {
	if m == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	cw := cbg.NewCborWriter(w)

	if _, err := cw.Write([]byte{161}); err != nil {
		return err
	}

	for _, k := range m.keys {
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
	*m = Map{}

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

var _ ipld.Map[string, any] = (*Map)(nil)
