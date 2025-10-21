package datamodel

import cbg "github.com/whyrusleeping/cbor-gen"

type MatchErrorModel struct {
	ErrorName string        `cborgen:"name"`
	Message   string        `cborgen:"message"`
	Statement *cbg.Deferred `cborgen:"statement,omitempty"`
	Cause     *cbg.Deferred `cborgen:"cause,omitempty"`
}

func (me MatchErrorModel) Name() string {
	return me.ErrorName
}

func (me MatchErrorModel) Error() string {
	return me.Message
}

var _ error = (*MatchErrorModel)(nil)
