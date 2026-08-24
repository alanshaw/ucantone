package promise

import (
	"github.com/ipfs/go-cid"
)

const (
	AwaitAnyTag   = "await/*"
	AwaitOKTag    = "await/ok"
	AwaitErrorTag = "await/error"
)

type AwaitAny struct {
	Task cid.Cid
}

type AwaitOK struct {
	Task cid.Cid
}

type AwaitError struct {
	Task cid.Cid
}
