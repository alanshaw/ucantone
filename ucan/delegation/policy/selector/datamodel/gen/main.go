package main

import (
	sdm "github.com/alanshaw/ucantone/ucan/delegation/policy/selector/datamodel"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func main() {
	if err := cbg.WriteMapEncodersToFile("../cbor_gen.go", "datamodel",
		sdm.ParseErrorModel{},
		sdm.ResolutionErrorModel{},
	); err != nil {
		panic(err)
	}
}
