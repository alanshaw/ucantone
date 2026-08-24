package datamodel

import (
	"github.com/fil-forge/ucantone/ipld/datamodel"
)

type PolicyModel struct {
	Statements []StatementModel `cborgen:"transparent" dagjsongen:"transparent"`
}

type StatementModel struct {
	Op         string            // Comparison, Wildcard, Conjunction, Disjunction, Negation, Quantification
	Selector   string            // Comparison, Quantification
	Statement  *StatementModel   // Negation
	Statements []*StatementModel // Conjunction, Disjunction, Quantification
	Pattern    string            // Wildcard
	Value      *datamodel.Any    // Comparison
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
	Op        string
	Selector  string
	Statement *StatementModel
}
