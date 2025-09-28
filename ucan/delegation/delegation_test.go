package delegation_test

import (
	"testing"

	"github.com/alanshaw/ucantone/testing/helpers"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/stretchr/testify/require"
)

func TestDelegation(t *testing.T) {
	t.Run("minimal", func(t *testing.T) {
		issuer := helpers.RandomSigner(t)
		audience := helpers.RandomDID(t)
		command := helpers.Must(command.Parse("/test/invoke"))(t)
		then := ucan.Now()

		initial, err := delegation.Delegate(issuer, audience, command)
		require.NoError(t, err)

		encoded, err := delegation.Encode(initial)
		require.NoError(t, err)

		decoded, err := delegation.Decode(encoded)
		require.NoError(t, err)

		require.Equal(t, issuer.DID(), decoded.Issuer().DID())
		require.Equal(t, audience, decoded.Audience())
		require.Equal(t, command, decoded.Command())
		require.Nil(t, decoded.Subject())
		require.NotEmpty(t, decoded.Nonce())
		require.GreaterOrEqual(t, *decoded.Expiration(), then)
	})
}
