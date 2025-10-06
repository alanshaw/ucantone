package datamodel

import (
	rdm "github.com/alanshaw/ucantone/result/datamodel"
	"github.com/ipfs/go-cid"
)

type ArgsModel struct {
	// Ran is the CID of the executed task the receipt is for.
	Ran cid.Cid `cborgen:"ran"`
	// Out is the attested result of the execution of the task.
	Out rdm.ResultModel `cborgen:"out"`
	// TODO: add Run
}
