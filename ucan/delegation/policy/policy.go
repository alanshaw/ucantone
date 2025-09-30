package policy

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/alanshaw/ucantone/ipld/datamodel"
	pdm "github.com/alanshaw/ucantone/ucan/delegation/policy/datamodel"
	"github.com/alanshaw/ucantone/ucan/delegation/policy/selector"
	"github.com/gobwas/glob"
	cbg "github.com/whyrusleeping/cbor-gen"
)

const (
	OpEqual              = "=="   // implemented by ComparisonStatement
	OpNotEqual           = "!="   // implemented by ComparisonStatement
	OpGreaterThan        = ">"    // implemented by ComparisonStatement
	OpGreaterThanOrEqual = ">="   // implemented by ComparisonStatement
	OpLessThan           = "<"    // implemented by ComparisonStatement
	OpLessThanOrEqual    = "<="   // implemented by ComparisonStatement
	OpAnd                = "and"  // implemented by ConjunctionStatement
	OpOr                 = "or"   // implemented by DisjunctionStatement
	OpNot                = "not"  // implemented by NegationStatement
	OpLike               = "like" // implemented by WildcardStatement
	OpAll                = "all"  // implemented by QuantificationStatement
	OpAny                = "any"  // implemented by QuantificationStatement
)

type Statement interface {
	Operation() string
}

// UCAN Delegation uses predicate logic statements extended with jq-inspired
// selectors as a policy language. Policies are syntactically driven, and
// constrain the args field of an eventual Invocation.
//
// https://github.com/ucan-wg/delegation/blob/main/README.md#policy
type Policy struct {
	// A Policy is always given as an array of predicates. This top-level array is
	// implicitly treated as a logical "and", where args MUST pass validation of
	// every top-level predicate.
	Statements []Statement
}

func (p Policy) MarshalCBOR(w io.Writer) error {
	var stmts []cbg.Deferred
	for _, s := range p.Statements {
		bytes, err := marshalCBORStatement(s)
		if err != nil {
			return err
		}
		stmts = append(stmts, cbg.Deferred{Raw: bytes})
	}
	model := pdm.PolicyModel{Statements: stmts}
	return model.MarshalCBOR(w)
}

func (p *Policy) UnmarshalCBOR(r io.Reader) error {
	var policyModel pdm.PolicyModel
	err := policyModel.UnmarshalCBOR(r)
	if err != nil {
		return err
	}
	for _, s := range policyModel.Statements {
		stmt, err := unmarshalCBORStatement(s.Raw)
		if err != nil {
			return err
		}
		p.Statements = append(p.Statements, stmt)
	}
	return nil
}

// https://github.com/ucan-wg/delegation/blob/main/README.md#comparisons
type ComparisonStatement struct {
	op       string
	Selector selector.Selector
	Value    any
}

func (cs ComparisonStatement) Operation() string {
	return cs.op
}

func (cs ComparisonStatement) MarshalCBOR(w io.Writer) error {
	model := pdm.ComparisonModel{
		Op:       cs.op,
		Selector: cs.Selector.String(),
		Value:    datamodel.NewAny(cs.Value),
	}
	return model.MarshalCBOR(w)
}

func (cs *ComparisonStatement) UnmarshalCBOR(r io.Reader) error {
	*cs = ComparisonStatement{}
	var model pdm.ComparisonModel
	err := model.UnmarshalCBOR(r)
	if err != nil {
		return err
	}
	if model.Op != OpEqual &&
		model.Op != OpNotEqual &&
		model.Op != OpGreaterThan &&
		model.Op != OpGreaterThanOrEqual &&
		model.Op != OpLessThan &&
		model.Op != OpLessThanOrEqual {
		return fmt.Errorf("unexpected comparison statement operation: %s", model.Op)
	}
	sel, err := selector.Parse(model.Selector)
	if err != nil {
		return err
	}
	cs.op = model.Op
	cs.Selector = sel
	cs.Value = model.Value.Value
	return nil
}

// https://github.com/ucan-wg/delegation/blob/main/README.md#connectives
type ConjunctionStatement struct {
	Statements []Statement
}

func (ConjunctionStatement) Operation() string {
	return OpAnd
}

func (cs ConjunctionStatement) MarshalCBOR(w io.Writer) error {
	policy := Policy(cs)
	var b bytes.Buffer
	err := policy.MarshalCBOR(&b)
	if err != nil {
		return err
	}
	model := pdm.ConjunctionModel{Op: OpAnd, Statements: cbg.Deferred{Raw: b.Bytes()}}
	return model.MarshalCBOR(w)
}

func (cs *ConjunctionStatement) UnmarshalCBOR(r io.Reader) error {
	*cs = ConjunctionStatement{}
	var model pdm.ConjunctionModel
	err := model.UnmarshalCBOR(r)
	if err != nil {
		return err
	}
	if model.Op != OpAnd {
		return fmt.Errorf("unexpected conjunction statement operation: %s", model.Op)
	}
	var policy Policy
	err = policy.UnmarshalCBOR(bytes.NewReader(model.Statements.Raw))
	if err != nil {
		return err
	}
	cs.Statements = policy.Statements
	return nil
}

// https://github.com/ucan-wg/delegation/blob/main/README.md#connectives
type DisjunctionStatement struct {
	Statements []Statement
}

func (DisjunctionStatement) Operation() string {
	return OpOr
}

func (ds DisjunctionStatement) MarshalCBOR(w io.Writer) error {
	policy := Policy(ds)
	var b bytes.Buffer
	err := policy.MarshalCBOR(&b)
	if err != nil {
		return err
	}
	model := pdm.DisjunctionModel{Op: OpOr, Statements: cbg.Deferred{Raw: b.Bytes()}}
	return model.MarshalCBOR(w)
}

func (ds *DisjunctionStatement) UnmarshalCBOR(r io.Reader) error {
	*ds = DisjunctionStatement{}
	var model pdm.DisjunctionModel
	err := model.UnmarshalCBOR(r)
	if err != nil {
		return err
	}
	if model.Op != OpOr {
		return fmt.Errorf("unexpected disjunction statement operation: %s", model.Op)
	}
	policy := Policy{}
	err = policy.UnmarshalCBOR(bytes.NewReader(model.Statements.Raw))
	if err != nil {
		return err
	}
	ds.Statements = policy.Statements
	return nil
}

// https://github.com/ucan-wg/delegation/blob/main/README.md#connectives
type NegationStatement struct {
	Statement Statement
}

func (NegationStatement) Operation() string {
	return OpNot
}

func (ns NegationStatement) MarshalCBOR(w io.Writer) error {
	bytes, err := marshalCBORStatement(ns.Statement)
	if err != nil {
		return err
	}
	model := pdm.NegationModel{Op: OpNot, Statement: cbg.Deferred{Raw: bytes}}
	return model.MarshalCBOR(w)
}

func (ns *NegationStatement) UnmarshalCBOR(r io.Reader) error {
	*ns = NegationStatement{}
	var model pdm.NegationModel
	err := model.UnmarshalCBOR(r)
	if err != nil {
		return err
	}
	if model.Op != OpNot {
		return fmt.Errorf("unexpected negation statement operation: %s", model.Op)
	}
	stmt, err := unmarshalCBORStatement(model.Statement.Raw)
	if err != nil {
		return err
	}
	ns.Statement = stmt
	return nil
}

// https://github.com/ucan-wg/delegation/blob/main/README.md#glob-matching
type WildcardStatement struct {
	Selector selector.Selector
	Pattern  string
	Glob     glob.Glob
}

func (WildcardStatement) Operation() string {
	return OpLike
}

func (ws WildcardStatement) MarshalCBOR(w io.Writer) error {
	model := pdm.WildcardModel{
		Op:       OpLike,
		Selector: ws.Selector.String(),
		Pattern:  ws.Pattern,
	}
	return model.MarshalCBOR(w)
}

func (ws *WildcardStatement) UnmarshalCBOR(r io.Reader) error {
	*ws = WildcardStatement{}
	var model pdm.WildcardModel
	err := model.UnmarshalCBOR(r)
	if err != nil {
		return err
	}
	if model.Op != OpLike {
		return fmt.Errorf("unexpected wildcard statement operation: %s", model.Op)
	}
	sel, err := selector.Parse(model.Selector)
	if err != nil {
		return err
	}
	glb, err := glob.Compile(model.Pattern)
	if err != nil {
		return err
	}
	ws.Selector = sel
	ws.Pattern = model.Pattern
	ws.Glob = glb
	return nil
}

// https://github.com/ucan-wg/delegation/blob/main/README.md#quantification
type QuantificationStatement struct {
	op         string
	Selector   selector.Selector
	Statements []Statement
}

func (qs QuantificationStatement) Operation() string {
	return qs.op
}

func (qs QuantificationStatement) MarshalCBOR(w io.Writer) error {
	policy := Policy{qs.Statements}
	var b bytes.Buffer
	err := policy.MarshalCBOR(&b)
	if err != nil {
		return err
	}
	model := pdm.QuantificationModel{
		Op:         qs.op,
		Selector:   qs.Selector.String(),
		Statements: cbg.Deferred{Raw: b.Bytes()},
	}
	return model.MarshalCBOR(w)
}

func (qs *QuantificationStatement) UnmarshalCBOR(r io.Reader) error {
	*qs = QuantificationStatement{}
	var model pdm.QuantificationModel
	err := model.UnmarshalCBOR(r)
	if err != nil {
		return err
	}
	if model.Op != OpAny && model.Op != OpAll {
		return fmt.Errorf("unexpected quantification statement operation: %s", model.Op)
	}
	sel, err := selector.Parse(model.Selector)
	if err != nil {
		return err
	}
	var policy Policy
	err = policy.UnmarshalCBOR(bytes.NewReader(model.Statements.Raw))
	if err != nil {
		return err
	}
	qs.Selector = sel
	qs.Statements = policy.Statements
	return nil
}

func Equal(selector selector.Selector, value any) ComparisonStatement {
	return ComparisonStatement{OpEqual, selector, value}
}

func GreaterThan(selector selector.Selector, value any) ComparisonStatement {
	return ComparisonStatement{OpGreaterThan, selector, value}
}

func GreaterThanOrEqual(selector selector.Selector, value any) ComparisonStatement {
	return ComparisonStatement{OpGreaterThanOrEqual, selector, value}
}

func LessThan(selector selector.Selector, value any) ComparisonStatement {
	return ComparisonStatement{OpLessThan, selector, value}
}

func LessThanOrEqual(selector selector.Selector, value any) ComparisonStatement {
	return ComparisonStatement{OpLessThanOrEqual, selector, value}
}

func Not(stmt Statement) NegationStatement {
	return NegationStatement{stmt}
}

func And(stmts ...Statement) ConjunctionStatement {
	return ConjunctionStatement{stmts}
}

func Or(stmts ...Statement) DisjunctionStatement {
	return DisjunctionStatement{stmts}
}

func Like(selector selector.Selector, pattern string, glob glob.Glob) WildcardStatement {
	return WildcardStatement{selector, pattern, glob}
}

func All(selector selector.Selector, stmts ...Statement) QuantificationStatement {
	return QuantificationStatement{OpAll, selector, stmts}
}

func Any(selector selector.Selector, stmts ...Statement) QuantificationStatement {
	return QuantificationStatement{OpAny, selector, stmts}
}

func marshalCBORStatement(stmt Statement) ([]byte, error) {
	cms, ok := stmt.(cbg.CBORMarshaler)
	if !ok {
		return nil, errors.New("statement is not CBOR marshaler")
	}
	var b bytes.Buffer
	err := cms.MarshalCBOR(&b)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func unmarshalCBORStatement(data []byte) (Statement, error) {
	var statementModel pdm.StatementModel
	// TODO: find a way to not read it twice
	err := statementModel.UnmarshalCBOR(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	var stmt cbg.CBORUnmarshaler
	switch statementModel.Op {
	case OpEqual, OpNotEqual, OpGreaterThan, OpGreaterThanOrEqual, OpLessThan, OpLessThanOrEqual:
		stmt = &ComparisonStatement{}
	case OpAnd:
		stmt = &ConjunctionStatement{}
	case OpOr:
		stmt = &DisjunctionStatement{}
	case OpNot:
		stmt = &NegationStatement{}
	case OpLike:
		stmt = &WildcardStatement{}
	case OpAny, OpAll:
		stmt = &QuantificationStatement{}
	}
	err = stmt.UnmarshalCBOR(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return stmt.(Statement), nil
}
