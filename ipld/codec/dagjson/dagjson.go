package dagjson

import (
	jsg "github.com/alanshaw/dag-json-gen"
)

const Code = 0x0129

type DagJsonMarshaler = jsg.DagJsonMarshaler
type DagJsonUnmarshaler = jsg.DagJsonUnmarshaler

type DagJsonMarshalable interface {
	DagJsonMarshaler
	DagJsonUnmarshaler
}
