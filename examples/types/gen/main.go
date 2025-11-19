package main

import (
	"github.com/alanshaw/ucantone/examples/types"
	cbg "github.com/whyrusleeping/cbor-gen"
)

func main() {
	if err := cbg.WriteMapEncodersToFile("../cbor_gen.go", "types",
		types.EmailsListArguments{},
		types.MessageSendArguments{},
		types.PromisedMsgSendArguments{},
	); err != nil {
		panic(err)
	}
}
