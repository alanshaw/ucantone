package main

import (
	cdm "github.com/alanshaw/ucantone/ucan/container/datamodel"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func main() {
	if err := cbg.WriteMapEncodersToFile("../cbor_gen.go", "datamodel",
		cdm.ContainerModel{},
	); err != nil {
		panic(err)
	}
}
