//go:generate go run -tags codegen .

package main

import (
	"os"

	jsg "github.com/alanshaw/dag-json-gen"
	idm "github.com/fil-forge/ucantone/ucan/invocation/datamodel"
	cbg "github.com/whyrusleeping/cbor-gen"
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
		idm.TaskModel{},
		idm.TokenPayloadModel1_0_0_rc1{},
		idm.SigPayloadModel{},
	}
	const (
		cborFile = "../cbor_gen.maps.go"
		jsonFile = "../dag_json_gen.maps.go"
	)
	if err := cbg.WriteMapEncodersToFile(cborFile, "datamodel", models...); err != nil {
		panic(err)
	}
	if err := jsg.WriteMapEncodersToFile(jsonFile, "datamodel", models...); err != nil {
		panic(err)
	}
	tag(cborFile)
	tag(jsonFile)
}
