package dispatcher_test

import (
	"fmt"
	"testing"

	"github.com/alanshaw/ucantone/errors"
	"github.com/alanshaw/ucantone/execution"
	"github.com/alanshaw/ucantone/execution/dispatcher"
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/testutil"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/alanshaw/ucantone/validator/capability"
	verrs "github.com/alanshaw/ucantone/validator/errors"
	"github.com/stretchr/testify/require"
)

func TestDispatcher(t *testing.T) {
	service := testutil.RandomSigner(t)
	alice := testutil.RandomSigner(t)

	// logs a message to the console
	consoleLogCapability, err := capability.New(
		"/console/log",
		capability.WithPolicyBuilder(policy.NotEqual(".message", "")),
	)
	require.NoError(t, err)

	// echos the arguments back to the caller
	testEchoCapability, err := capability.New("/test/echo")
	require.NoError(t, err)

	t.Run("dispatches invocations for execution", func(t *testing.T) {
		executor := dispatcher.New(service.Verifier())

		var messages []ipld.Any
		executor.Handle(consoleLogCapability, func(req execution.Request) (execution.Response, error) {
			msg := req.Invocation().Arguments()["message"]
			t.Log(msg)
			messages = append(messages, msg)
			return execution.NewResponse()
		})
		executor.Handle(testEchoCapability, func(req execution.Request) (execution.Response, error) {
			return execution.NewResponse(execution.WithSuccess(req.Invocation().Arguments()))
		})

		logInv, err := consoleLogCapability.Invoke(
			alice,
			alice,
			datamodel.Map{"message": "Hello, World!"},
			invocation.WithAudience(service),
		)
		require.NoError(t, err)

		resp, err := executor.Execute(execution.NewRequest(t.Context(), logInv, nil))
		require.NoError(t, err)

		_, x := result.Unwrap(resp.Result())
		require.Nil(t, x)

		require.Len(t, messages, 1)
		require.Equal(t, "Hello, World!", messages[0])

		echoInv, err := testEchoCapability.Invoke(
			alice,
			alice,
			datamodel.Map{"message": "echo!"},
			invocation.WithAudience(service),
		)
		require.NoError(t, err)

		resp, err = executor.Execute(execution.NewRequest(t.Context(), echoInv, nil))
		require.NoError(t, err)

		o, x := result.Unwrap(resp.Result())
		require.NotNil(t, o)
		require.Nil(t, x)
		t.Log(o)

		require.Len(t, messages, 1) // should not have changed
		require.Equal(t, "echo!", o.(ipld.Map)["message"])
	})

	t.Run("handler not found", func(t *testing.T) {
		executor := dispatcher.New(service.Verifier())

		inv, err := testEchoCapability.Invoke(
			alice,
			alice,
			datamodel.Map{"message": "echo!"},
			invocation.WithAudience(service),
		)
		require.NoError(t, err)

		resp, err := executor.Execute(execution.NewRequest(t.Context(), inv, nil))
		require.NoError(t, err)

		o, x := result.Unwrap(resp.Result())
		require.Nil(t, o)
		require.NotNil(t, x)
		t.Log(x)

		namedErr, ok := x.(errors.Named)
		require.True(t, ok)
		require.Equal(t, dispatcher.HandlerNotFoundErrorName, namedErr.Name())
	})

	t.Run("invalid audience", func(t *testing.T) {
		executor := dispatcher.New(service.Verifier())

		inv, err := testEchoCapability.Invoke(
			alice,
			alice,
			datamodel.Map{"message": "echo!"},
			invocation.WithAudience(alice),
		)
		require.NoError(t, err)

		resp, err := executor.Execute(execution.NewRequest(t.Context(), inv, nil))
		require.NoError(t, err)

		o, x := result.Unwrap(resp.Result())
		require.Nil(t, o)
		require.NotNil(t, x)
		t.Log(x)

		namedErr, ok := x.(errors.Named)
		require.True(t, ok)
		require.Equal(t, execution.InvalidAudienceErrorName, namedErr.Name())
	})

	t.Run("handler execution error", func(t *testing.T) {
		executor := dispatcher.New(service.Verifier())

		executor.Handle(consoleLogCapability, func(req execution.Request) (execution.Response, error) {
			return nil, fmt.Errorf("boom")
		})

		logInv, err := consoleLogCapability.Invoke(
			alice,
			alice,
			datamodel.Map{"message": "Hello, World!"},
			invocation.WithAudience(service),
		)
		require.NoError(t, err)

		resp, err := executor.Execute(execution.NewRequest(t.Context(), logInv, nil))
		require.NoError(t, err)

		o, x := result.Unwrap(resp.Result())
		require.Nil(t, o)
		require.NotNil(t, x)
		t.Log(x)

		namedErr, ok := x.(errors.Named)
		require.True(t, ok)
		require.Equal(t, execution.HandlerExecutionErrorName, namedErr.Name())
	})

	t.Run("validation error", func(t *testing.T) {
		executor := dispatcher.New(service.Verifier())
		executor.Handle(testEchoCapability, func(req execution.Request) (execution.Response, error) {
			return execution.NewResponse(execution.WithSuccess(req.Invocation().Arguments()))
		})

		logInv, err := testEchoCapability.Invoke(
			alice,
			testutil.RandomDID(t), // alice has no authority to invoke with this subject
			datamodel.Map{"message": "Hello, World!"},
			invocation.WithAudience(service),
		)
		require.NoError(t, err)

		resp, err := executor.Execute(execution.NewRequest(t.Context(), logInv, nil))
		require.NoError(t, err)

		o, x := result.Unwrap(resp.Result())
		require.Nil(t, o)
		require.NotNil(t, x)
		t.Log(x)

		namedErr, ok := x.(errors.Named)
		require.True(t, ok)
		require.Equal(t, verrs.InvalidClaimErrorName, namedErr.Name())
	})
}
