package main

import (
	rdm "github.com/alanshaw/ucantone/ucan/receipt/datamodel"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func main() {
	if err := cbg.WriteMapEncodersToFile("../cbor_gen.go", "datamodel",
		rdm.ArgsModel{},
	); err != nil {
		panic(err)
	}
}
