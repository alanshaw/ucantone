package datamodel

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
)

// Any is a CBOR backed implementation of the IPLD data model. Any supports
// serializing/deserializing the following kinds/types:
//
//   - Null (nil)
//   - Boolean (bool)
//   - Integer (int64)
//   - String (string)
//   - Bytes ([]byte)
//   - List (slice)
//   - Map ([Map])
//   - Link ([cid.Cid])
//
// Map values and list items may be any of the above types.
type Any struct {
	Value any
}

// NewAny creates an CBOR backed IPLD data model type from the passed data. The
// following Go types are supported:
//
//   - nil
//   - bool
//   - int64
//   - string
//   - []byte
//   - slice
//   - [Map]
//   - [cid.Cid]
func NewAny(data any) *Any {
	return &Any{Value: data}
}

func (a *Any) MarshalCBOR(w io.Writer) error {
	if a == nil || a.Value == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	switch v := a.Value.(type) {
	case *Map:
		return v.MarshalCBOR(w)
	case int64:
		return cbg.CborInt(v).MarshalCBOR(w)
	case bool:
		return cbg.CborBool(v).MarshalCBOR(w)
	case cid.Cid:
		return cbg.CborCid(v).MarshalCBOR(w)
	case string:
		cw := cbg.NewCborWriter(w)
		if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len(v))); err != nil {
			return err
		}
		_, err := cw.WriteString(v)
		return err
	case []byte:
		cw := cbg.NewCborWriter(w)
		if err := cw.WriteMajorTypeHeader(cbg.MajByteString, uint64(len(v))); err != nil {
			return err
		}
		_, err := cw.Write(v)
		return err
	}

	rt := reflect.TypeOf(a.Value)
	switch rt.Kind() {
	case reflect.Slice:
		cw := cbg.NewCborWriter(w)
		s := reflect.ValueOf(a.Value)
		if err := cw.WriteMajorTypeHeader(cbg.MajArray, uint64(s.Len())); err != nil {
			return err
		}
		for i := range s.Len() {
			a := Any{Value: s.Index(i).Interface()}
			if err := a.MarshalCBOR(w); err != nil {
				return fmt.Errorf("marshalling slice index: %d: %w", i, err)
			}
		}
		return nil
	}

	return fmt.Errorf("unsupported type: %T", a.Value)
}

func (a *Any) UnmarshalCBOR(r io.Reader) (err error) {
	*a = Any{}
	maj, extra, pr, err := peekCborHeader(r)
	if err != nil {
		return fmt.Errorf("peeking CBOR header: %w", err)
	}

	switch maj {
	case cbg.MajMap:
		var m Map
		a.Value = &m
		return m.UnmarshalCBOR(pr)
	case cbg.MajUnsignedInt, cbg.MajNegativeInt:
		var cbi cbg.CborInt
		if err = cbi.UnmarshalCBOR(pr); err != nil {
			return err
		}
		a.Value = int64(cbi)
		return nil
	case cbg.MajOther:
		switch extra {
		case 20:
			a.Value = false
			return nil
		case 21:
			a.Value = true
			return nil
		case 22: // null
			return nil
		}
	case cbg.MajTag:
		switch extra {
		case 42:
			cbc := cbg.CborCid{}
			if err = cbc.UnmarshalCBOR(pr); err != nil {
				return err
			}
			a.Value = cid.Cid(cbc)
			return nil
		}
	case cbg.MajTextString:
		if extra > 0 {
			cr := cbg.NewCborReader(pr)
			str, err := cbg.ReadStringWithMax(cr, cbg.MaxLength)
			if err != nil {
				return err
			}
			a.Value = str
		} else {
			a.Value = ""
		}
		return nil
	case cbg.MajByteString:
		if extra > 0 {
			cr := cbg.NewCborReader(pr)
			bytes, err := cbg.ReadByteArray(cr, cbg.ByteArrayMaxLen)
			if err != nil {
				return err
			}
			a.Value = bytes
		} else {
			a.Value = []byte{}
		}
		return nil
	case cbg.MajArray:
		if extra > cbg.MaxLength {
			return fmt.Errorf("array too large (%d)", extra)
		}
		if extra > 0 {
			var items reflect.Value
			for i := range extra {
				item := Any{}
				if err := item.UnmarshalCBOR(r); err != nil {
					return err
				}
				// TODO: ensure all items are the same type?
				if i == 0 {
					typ := reflect.TypeOf(item.Value)
					if typ == nil {
						return fmt.Errorf("nil item in list")
					}
					items = reflect.MakeSlice(reflect.SliceOf(typ), 0, int(extra))
				}
				items = reflect.Append(items, reflect.ValueOf(item.Value))
			}
			a.Value = items.Interface()
		} else {
			a.Value = []any{}
		}
		return nil
	}

	return fmt.Errorf("unsupported CBOR type: %d", maj)
}

func peekCborHeader(r io.Reader) (byte, uint64, io.Reader, error) {
	cr := cbg.NewCborReader(r)
	maj, extra, err := cr.ReadHeader()
	if err != nil {
		return 0, 0, nil, err
	}
	defer func() {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
	}()
	// TODO: find a better way of doing this
	var headerBuf bytes.Buffer
	cw := cbg.NewCborWriter(&headerBuf)
	err = cw.CborWriteHeader(maj, extra)
	if err != nil {
		return 0, 0, nil, err
	}
	return maj, extra, io.MultiReader(&headerBuf, r), nil
}
