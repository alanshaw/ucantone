package main

import (
	jsg "github.com/alanshaw/dag-json-gen"
	"github.com/alanshaw/ucantone/validator/internal/fixtures"
)

func main() {
	if err := jsg.WriteMapEncodersToFile("../dag_json_gen.go", "fixtures",
		fixtures.ErrorModel{},
		fixtures.FixturesModel{},
		fixtures.InvalidModel{},
		fixtures.ValidModel{},
	); err != nil {
		panic(err)
	}
}
