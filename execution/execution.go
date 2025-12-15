package execution

import (
	"context"

	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/ucan"
)

type Request interface {
	Context() context.Context
	// Invocation that should be executed.
	Invocation() ucan.Invocation
	// Metadata provides additional information about the invocation.
	Metadata() ucan.Container
}

type Response interface {
	// Result is the result of the task.
	Result() result.Result[ipld.Any, ipld.Any]
	// Metadata provides additional information about the response.
	Metadata() ucan.Container
}

// Executor executes UCAN invocations. In order to execute an invocation, proof
// chains must be validated and delegation policies matched. Hence a UCAN
// executor is responsible for both validation and execution of invocations.
type Executor interface {
	Execute(req Request) (Response, error)
}

// HandlerFunc is a function that can handle a specific UCAN invocation.
type HandlerFunc = func(req Request) (Response, error)
