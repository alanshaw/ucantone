//go:generate go run -tags codegen .

package main

import (
	"os"

	jsg "github.com/alanshaw/dag-json-gen"
	"github.com/fil-forge/ucantone/ucan/delegation/policy/selector/internal/fixtures/datamodel"
)

const buildTag = "//go:build !codegen\n\n"

func tag(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(path, append([]byte(buildTag), data...), 0644); err != nil {
		panic(err)
	}
}

func main() {
	models := []any{
		datamodel.FixtureModel{},
		datamodel.FixturesModel{},
	}
	const jsonFile = "../dag_json_gen.go"
	if err := jsg.WriteMapEncodersToFile(jsonFile, "datamodel", models...); err != nil {
		panic(err)
	}
	tag(jsonFile)
}
