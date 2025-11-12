package policy_test

import (
	"bytes"
	"testing"

	"github.com/alanshaw/ucantone/testutil"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/stretchr/testify/require"
)

func TestRoundtripCBOR(t *testing.T) {
	initial, err := policy.New(
		testutil.Must(policy.Equal(".foo", "bar"))(t),
	)
	require.NoError(t, err)
	var b bytes.Buffer
	err = initial.MarshalCBOR(&b)
	require.NoError(t, err)

	var decoded policy.Policy
	err = decoded.UnmarshalCBOR(&b)
	require.NoError(t, err)
	require.Len(t, decoded.Statements(), 1)
	require.Equal(t, policy.OpEqual, decoded.Statements()[0].Operator())
}

func TestParse(t *testing.T) {
	initial, err := policy.Parse(`[
		["==", ".foo", "bar"],
		["like", ".baz", "boz"]
	]`)
	require.NoError(t, err)
	require.Len(t, initial.Statements(), 2)
}
