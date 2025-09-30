package policy

import (
	"github.com/alanshaw/ucantone/ucan/delegation/policy/selector"
	"github.com/gobwas/glob"
)

const (
	KindEqual              = "=="   // implemented by EqualityStatement
	KindNotEqual           = "!="   // implemented by EqualityStatement
	KindGreaterThan        = ">"    // implemented by EqualityStatement
	KindGreaterThanOrEqual = ">="   // implemented by EqualityStatement
	KindLessThan           = "<"    // implemented by EqualityStatement
	KindLessThanOrEqual    = "<="   // implemented by EqualityStatement
	KindAnd                = "and"  // implemented by ConjunctionStatement
	KindOr                 = "or"   // implemented by DisjunctionStatement
	KindNot                = "not"  // implemented by NegationStatement
	KindLike               = "like" // implemented by WildcardStatement
	KindAll                = "all"  // implemented by QuantifierStatement
	KindAny                = "any"  // implemented by QuantifierStatement
)

type Policy = []Statement

type Statement interface {
	Kind() string
}

type EqualityStatement struct {
	kind     string
	Selector selector.Selector
	Value    any
}

func (es EqualityStatement) Kind() string {
	return es.kind
}

type ConjunctionStatement struct {
	Statements []Statement
}

func (ConjunctionStatement) Kind() string {
	return KindAnd
}

type DisjunctionStatement struct {
	Statements []Statement
}

func (DisjunctionStatement) Kind() string {
	return KindOr
}

type NegationStatement struct {
	Statement Statement
}

func (NegationStatement) Kind() string {
	return KindNot
}

type WildcardStatement struct {
	Selector selector.Selector
	Glob     glob.Glob
}

func (WildcardStatement) Kind() string {
	return KindLike
}

type QuantifierStatement struct {
	kind       string
	Selector   selector.Selector
	Statements []Statement
}

func (qs QuantifierStatement) Kind() string {
	return qs.kind
}

func Equal(selector selector.Selector, value any) EqualityStatement {
	return EqualityStatement{KindEqual, selector, value}
}

func GreaterThan(selector selector.Selector, value any) EqualityStatement {
	return EqualityStatement{KindGreaterThan, selector, value}
}

func GreaterThanOrEqual(selector selector.Selector, value any) EqualityStatement {
	return EqualityStatement{KindGreaterThanOrEqual, selector, value}
}

func LessThan(selector selector.Selector, value any) EqualityStatement {
	return EqualityStatement{KindLessThan, selector, value}
}

func LessThanOrEqual(selector selector.Selector, value any) EqualityStatement {
	return EqualityStatement{KindLessThanOrEqual, selector, value}
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

func Like(selector selector.Selector, glob glob.Glob) WildcardStatement {
	return WildcardStatement{selector, glob}
}

func All(selector selector.Selector, stmts ...Statement) QuantifierStatement {
	return QuantifierStatement{KindAll, selector, stmts}
}

func Any(selector selector.Selector, stmts ...Statement) QuantifierStatement {
	return QuantifierStatement{KindAny, selector, stmts}
}
