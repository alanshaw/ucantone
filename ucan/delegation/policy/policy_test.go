package policy_test

import (
	"bytes"
	"testing"

	"github.com/alanshaw/ucantone/testutil"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/alanshaw/ucantone/ucan/delegation/policy/selector"
	"github.com/stretchr/testify/require"
)

func TestRoundtripCBOR(t *testing.T) {
	initial := policy.Policy{
		Statements: []policy.Statement{
			policy.Equal(testutil.Must(selector.Parse(".foo"))(t), "bar"),
		},
	}
	var b bytes.Buffer
	err := initial.MarshalCBOR(&b)
	require.NoError(t, err)

	var decoded policy.Policy
	err = decoded.UnmarshalCBOR(&b)
	require.NoError(t, err)
	require.Len(t, decoded.Statements, 1)
	require.Equal(t, policy.OpEqual, decoded.Statements[0].Operation())
}
