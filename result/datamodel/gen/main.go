package main

import (
	rdm "github.com/alanshaw/ucantone/result/datamodel"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func main() {
	if err := cbg.WriteMapEncodersToFile("../cbor_gen.go", "datamodel",
		rdm.ResultModel{},
	); err != nil {
		panic(err)
	}
}
