package main

import (
	idm "github.com/alanshaw/ucantone/ucan/invocation/datamodel"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func main() {
	if err := cbg.WriteTupleEncodersToFile("../envelope_cbor_gen.go", "datamodel",
		idm.EnvelopeModel{},
	); err != nil {
		panic(err)
	}
	if err := cbg.WriteMapEncodersToFile("../invocation_cbor_gen.go", "datamodel",
		idm.TokenPayloadModel1_0_0_rc1{},
		idm.SigPayloadModel{},
	); err != nil {
		panic(err)
	}
}
