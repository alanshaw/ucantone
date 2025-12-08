package main

import (
	vdm "github.com/alanshaw/ucantone/errors/datamodel"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func main() {
	if err := cbg.WriteMapEncodersToFile("../cbor_gen.go", "datamodel",
		vdm.ErrorModel{},
	); err != nil {
		panic(err)
	}
}
