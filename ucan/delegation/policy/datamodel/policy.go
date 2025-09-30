package datamodel

import (
	"github.com/alanshaw/ucantone/ipld/datamodel"
	cbg "github.com/whyrusleeping/cbor-gen"
)

type PolicyModel struct {
	Statements []cbg.Deferred `cborgen:"transparent"`
}

type StatementModel struct {
	Op   string
	Arg0 cbg.Deferred `cborgen:"optional"`
	Arg1 cbg.Deferred `cborgen:"optional"`
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
	Statements cbg.Deferred
}

type DisjunctionModel struct {
	Op         string
	Statements cbg.Deferred
}

type NegationModel struct {
	Op        string
	Statement cbg.Deferred
}

type QuantificationModel struct {
	Op         string
	Selector   string
	Statements cbg.Deferred
}
