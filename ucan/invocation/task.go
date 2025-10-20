package invocation

import (
	"bytes"
	"fmt"

	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/ucan"
	idm "github.com/alanshaw/ucantone/ucan/invocation/datamodel"
	cid "github.com/ipfs/go-cid"
	multihash "github.com/multiformats/go-multihash/core"
	cbg "github.com/whyrusleeping/cbor-gen"
)

type Task struct {
	link  cid.Cid
	bytes []byte
	sub   ucan.Subject
	cmd   ucan.Command
	args  ipld.Map[string, ipld.Any]
	nnc   ucan.Nonce
}

func NewTask(
	subject ucan.Subject,
	command ucan.Command,
	arguments ipld.Map[string, ipld.Any],
	nonce ucan.Nonce,
) (*Task, error) {
	var args cbg.Deferred
	argsMap := datamodel.NewMap(datamodel.WithEntries(arguments.Entries()))
	var argsBuf bytes.Buffer
	err := argsMap.MarshalCBOR(&argsBuf)
	if err != nil {
		return nil, fmt.Errorf("marshaling arguments CBOR: %w", err)
	}
	args.Raw = argsBuf.Bytes()

	taskModel := idm.TaskModel{
		Sub:   subject.DID(),
		Cmd:   command,
		Args:  args,
		Nonce: nonce,
	}
	var taskBuf bytes.Buffer
	err = taskModel.MarshalCBOR(&taskBuf)
	if err != nil {
		return nil, fmt.Errorf("marshaling task CBOR: %w", err)
	}
	link, err := cid.V1Builder{
		Codec:  dagcbor.Code,
		MhType: multihash.SHA2_256,
	}.Sum(taskBuf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("hashing task bytes: %w", err)
	}

	return &Task{link, taskBuf.Bytes(), subject, command, arguments, nonce}, nil
}

func (t *Task) Arguments() ipld.Map[string, ipld.Any] {
	return t.args
}

func (t *Task) Bytes() []byte {
	return t.bytes
}

func (t *Task) Command() ucan.Command {
	return t.cmd
}

func (t *Task) Link() cid.Cid {
	return t.link
}

func (t *Task) Nonce() ucan.Nonce {
	return t.nnc
}

func (t *Task) Subject() ucan.Subject {
	return t.sub
}

var _ ucan.Task = (*Task)(nil)
