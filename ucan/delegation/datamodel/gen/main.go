package main

import (
	ddm "github.com/alanshaw/ucantone/ucan/delegation/datamodel"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func main() {
	if err := cbg.WriteTupleEncodersToFile("../cbor_gen.tuples.go", "datamodel",
		ddm.EnvelopeModel{},
	); err != nil {
		panic(err)
	}
	if err := cbg.WriteMapEncodersToFile("../cbor_gen.maps.go", "datamodel",
		ddm.TokenPayloadModel1_0_0_rc1{},
		ddm.SigPayloadModel{},
	); err != nil {
		panic(err)
	}
}
