package bindexec_test

import (
	"testing"

	"github.com/alanshaw/ucantone/execution"
	"github.com/alanshaw/ucantone/execution/bindexec"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/result"
	"github.com/alanshaw/ucantone/testutil"
	tdm "github.com/alanshaw/ucantone/testutil/datamodel"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	alice := testutil.RandomSigner(t)
	handler := bindexec.NewHandler(func(req *bindexec.Request[*tdm.TestObject], res *bindexec.Response[*tdm.TestObject2]) error {
		args := req.Task().BindArguments()
		require.Equal(t, args.Bytes, []byte{0x01, 0x02, 0x03})
		return res.SetSuccess(&tdm.TestObject2{Str: "testy"})
	})

	inv, err := invocation.Invoke(
		alice,
		alice,
		"/test/handler",
		datamodel.Map{"bytes": []byte{0x01, 0x02, 0x03}},
	)
	require.NoError(t, err)

	req := execution.NewRequest(t.Context(), inv)
	require.NoError(t, err)

	res, err := execution.NewResponse(inv.Task().Link(), execution.WithSigner(alice))
	require.NoError(t, err)

	err = handler(req, res)
	require.NoError(t, err)
	require.NotNil(t, res)

	o, x := result.Unwrap(res.Receipt().Out())
	require.Nil(t, x)
	require.NotNil(t, o)
	require.Equal(t, "testy", o.(datamodel.Map)["str"])
}
