package main

import (
	"github.com/alanshaw/ucantone/testing/helpers"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func main() {
	if err := cbg.WriteMapEncodersToFile("../cbor_gen.go", "helpers",
		helpers.TestObject{},
		helpers.TestArgs{},
	); err != nil {
		panic(err)
	}
}
