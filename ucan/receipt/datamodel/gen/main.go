package main

import (
	rdm "github.com/alanshaw/ucantone/ucan/receipt/datamodel"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func main() {
	if err := cbg.WriteTupleEncodersToFile("../cbor_gen.tuples.go", "datamodel",
		rdm.EnvelopeModel{},
	); err != nil {
		panic(err)
	}
	if err := cbg.WriteMapEncodersToFile("../cbor_gen.maps.go", "datamodel",
		rdm.ArgsModel{},
		rdm.TokenPayloadModel1_0_0_rc1{},
		rdm.SigPayloadModel{},
	); err != nil {
		panic(err)
	}
}
