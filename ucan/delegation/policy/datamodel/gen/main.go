package main

import (
	pdm "github.com/alanshaw/ucantone/ucan/delegation/policy/datamodel"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func main() {
	if err := cbg.WriteTupleEncodersToFile("../cbor_gen.go", "datamodel",
		pdm.PolicyModel{},
		pdm.StatementModel{},
		pdm.ComparisonModel{},
		pdm.WildcardModel{},
		pdm.ConjunctionModel{},
		pdm.DisjunctionModel{},
		pdm.NegationModel{},
		pdm.QuantificationModel{},
	); err != nil {
		panic(err)
	}
}
