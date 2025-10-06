package helpers

import (
	"testing"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/require"
)

type TestObject struct {
	Bytes []byte `cborgen:"bytes"`
}

type TestObject2 struct {
	Str   string `cborgen:"str"`
	Bytes []byte `cborgen:"bytes"`
}

type TestArgs struct {
	ID    did.DID     `cborgen:"id"`
	Link  cid.Cid     `cborgen:"link"`
	Str   string      `cborgen:"str"`
	Num   int64       `cborgen:"num"`
	Bytes []byte      `cborgen:"bytes"`
	Obj   TestObject  `cborgen:"obj"`
	Obj2  TestObject2 `cborgen:"obj2"`
	List  []string    `cborgen:"list"`
}

func RandomArgs(t *testing.T) ipld.Map[string, any] {
	var list []string
	for range RandomBytes(t, 1)[0] {
		list = append(list, RandomCID(t).String())
	}
	var m datamodel.Map
	err := datamodel.Rebind(&TestArgs{
		ID:    RandomDID(t),
		Link:  RandomCID(t),
		Str:   RandomCID(t).String(),
		Num:   int64(RandomBytes(t, 1)[0]),
		Bytes: RandomBytes(t, 32),
		Obj: TestObject{
			Bytes: RandomBytes(t, 32),
		},
		List: list,
	}, &m)
	require.NoError(t, err)
	return &m
}
