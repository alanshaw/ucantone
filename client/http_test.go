package client_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/alanshaw/ucantone/client"
	"github.com/alanshaw/ucantone/execution"
	"github.com/alanshaw/ucantone/ipld"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/server"
	"github.com/alanshaw/ucantone/testutil"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/stretchr/testify/require"
)

func TestHTTPClient(t *testing.T) {
	service := testutil.RandomSigner(t)
	alice := testutil.RandomSigner(t)

	t.Run("invocation execution round trip", func(t *testing.T) {
		server := server.NewHTTP(service)

		server.Handle(testutil.TestEchoCapability, func(req execution.Request) (execution.Response, error) {
			return execution.NewResponse(execution.WithSuccess(req.Invocation().Arguments()))
		})

		c, err := client.NewHTTP(
			testutil.Must(url.Parse("http://localhost"))(t),
			client.WithHTTPClient(&http.Client{Transport: server}),
		)
		require.NoError(t, err)

		inv, err := testutil.TestEchoCapability.Invoke(
			alice,
			alice,
			datamodel.Map{"message": "echo!"},
			invocation.WithAudience(service),
		)
		require.NoError(t, err)

		res, err := c.Execute(execution.NewRequest(t.Context(), inv))
		require.NoError(t, err)

		o, x := result.Unwrap(res.Result())
		require.Nil(t, x)
		require.NotNil(t, o)
		require.Equal(t, "echo!", o.(ipld.Map)["message"])
	})
}
