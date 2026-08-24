//go:generate go run -tags codegen .

package main

import (
	"os"

	"github.com/fil-forge/ucantone/examples/types"
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
		types.EmailsListArguments{},
		types.MessageSendArguments{},
		types.PromisedMsgSendArguments{},
		types.EchoArguments{},
	}
	const cborFile = "../cbor_gen.go"
	if err := cbg.WriteMapEncodersToFile(cborFile, "types", models...); err != nil {
		panic(err)
	}
	tag(cborFile)
}
