//go:build !codegen

package policy

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	pdm "github.com/fil-forge/ucantone/ucan/delegation/policy/datamodel"
)

func (p Policy) MarshalCBOR(w io.Writer) error {
	statements := make([]pdm.StatementModel, 0, len(p.statements))
	for _, s := range p.statements {
		statements = append(statements, s.model)
	}
	model := pdm.PolicyModel{Statements: statements}
	return model.MarshalCBOR(w)
}

func (p *Policy) UnmarshalCBOR(r io.Reader) error {
	*p = Policy{}
	var policyModel pdm.PolicyModel
	err := policyModel.UnmarshalCBOR(r)
	if err != nil {
		return err
	}
	for i, m := range policyModel.Statements {
		s, err := newStatement(m)
		if err != nil {
			return fmt.Errorf(`unmarshaling policy statement %d with operator "%s": %w`, i, m.Op, err)
		}
		p.statements = append(p.statements, s)
	}
	return nil
}

func (p Policy) MarshalDagJSON(w io.Writer) error {
	statements := make([]pdm.StatementModel, 0, len(p.statements))
	for _, s := range p.statements {
		statements = append(statements, s.model)
	}
	model := pdm.PolicyModel{Statements: statements}
	return model.MarshalDagJSON(w)
}

func (p *Policy) UnmarshalDagJSON(r io.Reader) error {
	*p = Policy{}
	var policyModel pdm.PolicyModel
	err := policyModel.UnmarshalDagJSON(r)
	if err != nil {
		return err
	}
	for i, m := range policyModel.Statements {
		s, err := newStatement(m)
		if err != nil {
			return fmt.Errorf(`unmarshaling policy statement %d with operator "%s": %w`, i, m.Op, err)
		}
		p.statements = append(p.statements, s)
	}
	return nil
}

func (p Policy) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	err := p.MarshalDagJSON(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p *Policy) UnmarshalJSON(b []byte) error {
	return p.UnmarshalDagJSON(bytes.NewReader(b))
}

func (p Policy) String() string {
	data, err := p.MarshalJSON()
	if err != nil {
		return fmt.Sprintf("Error: marshaling policy to string: %s", err.Error())
	}
	return string(data)
}

func (s Statement) MarshalCBOR(w io.Writer) error {
	return s.model.MarshalCBOR(w)
}

func (s *Statement) UnmarshalCBOR(r io.Reader) error {
	model := pdm.StatementModel{}
	if err := model.UnmarshalCBOR(r); err != nil {
		return err
	}
	stmt, err := newStatement(model)
	if err != nil {
		return err
	}
	*s = stmt
	return nil
}

func (s Statement) MarshalDagJSON(w io.Writer) error {
	return s.model.MarshalDagJSON(w)
}

func (s *Statement) UnmarshalDagJSON(r io.Reader) error {
	model := pdm.StatementModel{}
	if err := model.UnmarshalDagJSON(r); err != nil {
		return err
	}
	stmt, err := newStatement(model)
	if err != nil {
		return err
	}
	*s = stmt
	return nil
}

// Parse a policy encoded as a DAG-JSON string.
func Parse(input string) (Policy, error) {
	pol := Policy{}
	err := pol.UnmarshalDagJSON(strings.NewReader(input))
	if err != nil {
		return Policy{}, err
	}
	return pol, nil
}
