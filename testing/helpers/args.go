package helpers

import (
	"testing"

	"github.com/alanshaw/ucantone/did"
	"github.com/ipfs/go-cid"
)

type TestObject struct {
	Bytes []byte `cborgen:"bytes"`
}

type TestArgs struct {
	ID    did.DID    `cborgen:"id"`
	Link  cid.Cid    `cborgen:"link"`
	Str   string     `cborgen:"str"`
	Num   int64      `cborgen:"num"`
	Bytes []byte     `cborgen:"bytes"`
	Obj   TestObject `cborgen:"obj"`
	List  []string   `cborgen:"list"`
}

func RandomArgs(t *testing.T) *TestArgs {
	var list []string
	for range RandomBytes(t, 1)[0] {
		list = append(list, RandomCID(t).String())
	}
	return &TestArgs{
		ID:    RandomDID(t),
		Link:  RandomCID(t),
		Str:   RandomCID(t).String(),
		Num:   int64(RandomBytes(t, 1)[0]),
		Bytes: RandomBytes(t, 32),
		Obj: TestObject{
			Bytes: RandomBytes(t, 32),
		},
		List: list,
	}
}
