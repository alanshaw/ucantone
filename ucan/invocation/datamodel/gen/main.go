package main

import (
	idm "github.com/alanshaw/ucantone/ucan/invocation/datamodel"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func main() {
	if err := cbg.WriteTupleEncodersToFile("../cbor_gen.tuples.go", "datamodel",
		idm.EnvelopeModel{},
	); err != nil {
		panic(err)
	}
	if err := cbg.WriteMapEncodersToFile("../cbor_gen.maps.go", "datamodel",
		idm.TaskModel{},
		idm.TokenPayloadModel1_0_0_rc1{},
		idm.SigPayloadModel{},
	); err != nil {
		panic(err)
	}
}
