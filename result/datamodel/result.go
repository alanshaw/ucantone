package datamodel

import cbg "github.com/whyrusleeping/cbor-gen"

type ResultModel struct {
	Ok  *cbg.Deferred `cborgen:"ok,omitempty"`
	Err *cbg.Deferred `cborgen:"error,omitempty"`
}
