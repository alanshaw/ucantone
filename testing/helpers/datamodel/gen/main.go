package main

import (
	"github.com/alanshaw/ucantone/testing/helpers/datamodel"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func main() {
	if err := cbg.WriteMapEncodersToFile("../cbor_gen.go", "datamodel",
		datamodel.TestObject{},
		datamodel.TestObject2{},
		datamodel.TestArgs{},
	); err != nil {
		panic(err)
	}
}
