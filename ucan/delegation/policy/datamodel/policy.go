package datamodel

import (
	"bytes"
	"fmt"
	"io"

	"github.com/alanshaw/ucantone/ipld/datamodel"
	cbg "github.com/whyrusleeping/cbor-gen"
)

type PolicyModel struct {
	Statements []StatementModel `cborgen:"transparent"`
}

type StatementModel struct {
	Op         string            // Comparison, Wildcard, Conjunction, Disjunction, Negation, Quantification
	Selector   string            // Comparison, Quantification
	Statement  *StatementModel   // Negation
	Statements []*StatementModel // Conjunction, Disjunction, Quantification
	Pattern    string            // Wildcard
	Value      *datamodel.Any    // Comparison
}

func (t *StatementModel) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}

	switch t.Op {
	case "==", "!=", ">", ">=", "<", "<=":
		m := ComparisonModel{t.Op, t.Selector, t.Value}
		if err := m.MarshalCBOR(w); err != nil {
			return err
		}
	case "and":
		m := ConjunctionModel{t.Op, t.Statements}
		if err := m.MarshalCBOR(w); err != nil {
			return err
		}
	case "or":
		m := DisjunctionModel{t.Op, t.Statements}
		if err := m.MarshalCBOR(w); err != nil {
			return err
		}
	case "not":
		m := NegationModel{t.Op, t.Statement}
		if err := m.MarshalCBOR(w); err != nil {
			return err
		}
	case "like":
		m := WildcardModel{t.Op, t.Selector, t.Pattern}
		if err := m.MarshalCBOR(w); err != nil {
			return err
		}
	case "all", "any":
		m := QuantificationModel{t.Op, t.Selector, t.Statements}
		if err := m.MarshalCBOR(w); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown statement: %s", t.Op)
	}
	return nil
}

func (t *StatementModel) UnmarshalCBOR(r io.Reader) error {
	*t = StatementModel{}

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

	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra > 3 {
		return fmt.Errorf("cbor input has too many fields %d > 3", extra)
	}

	if extra < 1 {
		return fmt.Errorf("cbor input has too few fields %d < 1", extra)
	}

	op, err := cbg.ReadStringWithMax(cr, cbg.MaxLength)
	if err != nil {
		return err
	}
	t.Op = op

	switch op {
	case "==", "!=", ">", ">=", "<", "<=":
		// TODO: can probably do this upfront for each type
		b, err := encodeStatementBegin(3, op) // TODO: remove the magic number (number of fields)
		if err != nil {
			return err
		}
		m := ComparisonModel{}
		if err := m.UnmarshalCBOR(io.MultiReader(bytes.NewReader(b), cr)); err != nil {
			return err
		}
		t.Selector = m.Selector
		t.Value = m.Value
	case "and":
		// TODO: can probably do this upfront for each type
		b, err := encodeStatementBegin(2, op) // TODO: remove the magic number (number of fields)
		if err != nil {
			return err
		}
		m := ConjunctionModel{}
		if err := m.UnmarshalCBOR(io.MultiReader(bytes.NewReader(b), cr)); err != nil {
			return err
		}
		t.Statements = m.Statements
	case "or":
		// TODO: can probably do this upfront for each type
		b, err := encodeStatementBegin(2, op) // TODO: remove the magic number (number of fields)
		if err != nil {
			return err
		}
		m := DisjunctionModel{}
		if err := m.UnmarshalCBOR(io.MultiReader(bytes.NewReader(b), cr)); err != nil {
			return err
		}
		t.Statements = m.Statements
	case "not":
		// TODO: can probably do this upfront for each type
		b, err := encodeStatementBegin(2, op) // TODO: remove the magic number (number of fields)
		if err != nil {
			return err
		}
		m := NegationModel{}
		if err := m.UnmarshalCBOR(io.MultiReader(bytes.NewReader(b), cr)); err != nil {
			return err
		}
		t.Statement = m.Statement
	case "like":
		// TODO: can probably do this upfront for each type
		b, err := encodeStatementBegin(3, op) // TODO: remove the magic number (number of fields)
		if err != nil {
			return err
		}
		m := WildcardModel{}
		if err := m.UnmarshalCBOR(io.MultiReader(bytes.NewReader(b), cr)); err != nil {
			return err
		}
		t.Selector = m.Selector
		t.Pattern = m.Pattern
	case "all", "any":
		// TODO: can probably do this upfront for each type
		b, err := encodeStatementBegin(3, op) // TODO: remove the magic number (number of fields)
		if err != nil {
			return err
		}
		m := QuantificationModel{}
		if err := m.UnmarshalCBOR(io.MultiReader(bytes.NewReader(b), cr)); err != nil {
			return err
		}
		t.Selector = m.Selector
		t.Statements = m.Statements
	default:
		return fmt.Errorf("unknown statement: %s", t.Op)
	}
	return nil
}

func (t *StatementModel) MarshalDagJSON(w io.Writer) error {
	panic("not implemented")
}

func (t *StatementModel) UnmarshalDagJSON(r io.Reader) error {
	panic("not implemented")
}

func encodeStatementBegin(numFields uint64, op string) ([]byte, error) {
	var buf bytes.Buffer
	cw := cbg.NewCborWriter(&buf)
	if _, err := cw.Write(cbg.CborEncodeMajorType(cbg.MajArray, numFields)); err != nil {
		return nil, err
	}
	if err := cw.WriteMajorTypeHeader(cbg.MajTextString, uint64(len(op))); err != nil {
		return nil, err
	}
	if _, err := cw.WriteString(op); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type ComparisonModel struct {
	Op       string
	Selector string
	Value    *datamodel.Any
}

type WildcardModel struct {
	Op       string
	Selector string
	Pattern  string
}

type ConjunctionModel struct {
	Op         string
	Statements []*StatementModel
}

type DisjunctionModel struct {
	Op         string
	Statements []*StatementModel
}

type NegationModel struct {
	Op        string
	Statement *StatementModel
}

type QuantificationModel struct {
	Op         string
	Selector   string
	Statements []*StatementModel
}
