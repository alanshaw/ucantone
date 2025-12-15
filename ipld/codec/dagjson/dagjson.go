package dagjson

import (
	jsg "github.com/alanshaw/dag-json-gen"
)

const (
	Code        = 0x0129
	ContentType = "application/vnd.ipld.dag-json"
)

type DagJsonMarshaler = jsg.DagJsonMarshaler
type DagJsonUnmarshaler = jsg.DagJsonUnmarshaler

type DagJsonMarshalable interface {
	DagJsonMarshaler
	DagJsonUnmarshaler
}
