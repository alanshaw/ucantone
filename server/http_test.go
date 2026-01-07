package server_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/alanshaw/ucantone/execution"
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/server"
	"github.com/alanshaw/ucantone/testutil"
	"github.com/alanshaw/ucantone/ucan/container"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/stretchr/testify/require"
)

func TestHTTPServer(t *testing.T) {
	service := testutil.RandomSigner(t)
	alice := testutil.RandomSigner(t)

	t.Run("invocation execution round trip", func(t *testing.T) {
		server := server.NewHTTP(service)

		var messages []ipld.Any
		server.Handle(testutil.ConsoleLogCapability, func(req execution.Request) (execution.Response, error) {
			msg := req.Invocation().Arguments()["message"]
			t.Log(msg)
			messages = append(messages, msg)
			return execution.NewResponse(execution.WithSuccess(service, req.Invocation().Task().Link(), ipld.Map{}))
		})
		server.Handle(testutil.TestEchoCapability, func(req execution.Request) (execution.Response, error) {
			inv := req.Invocation()
			return execution.NewResponse(execution.WithSuccess(service, inv.Task().Link(), inv.Arguments()))
		})

		logInv, err := testutil.ConsoleLogCapability.Invoke(
			alice,
			alice,
			datamodel.Map{"message": "Hello, World!"},
			invocation.WithAudience(service),
		)
		require.NoError(t, err)

		ct := container.New(container.WithInvocations(logInv))

		r, w := io.Pipe()
		go func() {
			err := ct.MarshalCBOR(w)
			w.CloseWithError(err)
		}()

		req := http.Request{Header: http.Header{}, Body: r}
		req.Header.Set("Content-Type", dagcbor.ContentType)

		resp, err := server.RoundTrip(&req)
		require.NoError(t, err)

		ctResp := container.Container{}
		err = ctResp.UnmarshalCBOR(resp.Body)
		require.NoError(t, err)

		require.Len(t, ctResp.Receipts(), 1)

		_, x := result.Unwrap(ctResp.Receipts()[0].Out())
		require.Nil(t, x)

		require.Len(t, messages, 1)
		require.Equal(t, "Hello, World!", messages[0])

		echoInv, err := testutil.TestEchoCapability.Invoke(
			alice,
			alice,
			datamodel.Map{"message": "echo!"},
			invocation.WithAudience(service),
		)
		require.NoError(t, err)

		ct = container.New(container.WithInvocations(echoInv))

		r, w = io.Pipe()
		go func() {
			err := ct.MarshalCBOR(w)
			w.CloseWithError(err)
		}()

		req = http.Request{Header: http.Header{}, Body: r}
		req.Header.Set("Content-Type", dagcbor.ContentType)

		resp, err = server.RoundTrip(&req)
		require.NoError(t, err)

		ctResp = container.Container{}
		err = ctResp.UnmarshalCBOR(resp.Body)
		require.NoError(t, err)

		require.Len(t, ctResp.Receipts(), 1)

		o, x := result.Unwrap(ctResp.Receipts()[0].Out())
		require.NotNil(t, o)
		require.Nil(t, x)
		t.Log(o)

		require.Len(t, messages, 1) // should not have changed
		require.Equal(t, "echo!", o.(ipld.Map)["message"])
	})
}
