package invocation_test

import (
	"testing"

	"github.com/alanshaw/ucantone/testing/helpers"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/stretchr/testify/require"
)

func TestInvoke(t *testing.T) {
	t.Run("minimal", func(t *testing.T) {
		issuer := helpers.RandomSigner(t)
		subject := helpers.RandomDID(t)
		command := helpers.Must(command.Parse("/test/invoke"))(t)
		arguments := invocation.NoArguments
		then := ucan.Now()

		initial, err := invocation.Invoke(issuer, subject, command, arguments)
		require.NoError(t, err)

		encoded, err := invocation.Encode(initial)
		require.NoError(t, err)

		decoded, err := invocation.Decode(encoded)
		require.NoError(t, err)

		require.Equal(t, issuer.DID(), decoded.Issuer().DID())
		require.Equal(t, subject, decoded.Subject())
		require.Equal(t, command, decoded.Command())
		require.Nil(t, decoded.Audience())
		require.NotEmpty(t, decoded.Nonce())
		require.GreaterOrEqual(t, *decoded.Expiration(), then)
	})

	t.Run("bad command", func(t *testing.T) {
		issuer := helpers.RandomSigner(t)
		subject := helpers.RandomDID(t)
		arguments := invocation.NoArguments

		_, err := invocation.Invoke(issuer, subject, "testinvoke", arguments)
		require.Error(t, err)
		require.ErrorIs(t, err, command.ErrRequiresLeadingSlash)
	})

	t.Run("no nonce", func(t *testing.T) {
		issuer := helpers.RandomSigner(t)
		subject := helpers.RandomDID(t)
		command := helpers.Must(command.Parse("/test/invoke"))(t)
		arguments := invocation.NoArguments

		initial, err := invocation.Invoke(issuer, subject, command, arguments, invocation.WithNoNonce())
		require.NoError(t, err)

		encoded, err := invocation.Encode(initial)
		require.NoError(t, err)

		decoded, err := invocation.Decode(encoded)
		require.NoError(t, err)

		require.NoError(t, err)
		require.Nil(t, decoded.Nonce())
	})

	t.Run("custom nonce", func(t *testing.T) {
		issuer := helpers.RandomSigner(t)
		subject := helpers.RandomDID(t)
		command := helpers.Must(command.Parse("/test/invoke"))(t)
		arguments := invocation.NoArguments
		nonce := []byte{1, 2, 3}

		initial, err := invocation.Invoke(issuer, subject, command, arguments, invocation.WithNonce(nonce))
		require.NoError(t, err)

		encoded, err := invocation.Encode(initial)
		require.NoError(t, err)

		decoded, err := invocation.Decode(encoded)
		require.NoError(t, err)

		require.Equal(t, nonce, decoded.Nonce())
	})

	t.Run("no expiration", func(t *testing.T) {
		issuer := helpers.RandomSigner(t)
		subject := helpers.RandomDID(t)
		command := helpers.Must(command.Parse("/test/invoke"))(t)
		arguments := invocation.NoArguments

		initial, err := invocation.Invoke(issuer, subject, command, arguments, invocation.WithNoExpiration())
		require.NoError(t, err)

		encoded, err := invocation.Encode(initial)
		require.NoError(t, err)

		decoded, err := invocation.Decode(encoded)
		require.NoError(t, err)

		require.NoError(t, err)
		require.Nil(t, decoded.Expiration())
	})

	t.Run("custom expiration", func(t *testing.T) {
		issuer := helpers.RandomSigner(t)
		subject := helpers.RandomDID(t)
		command := helpers.Must(command.Parse("/test/invoke"))(t)
		arguments := invocation.NoArguments
		expiration := ucan.Now() + 138

		initial, err := invocation.Invoke(issuer, subject, command, arguments, invocation.WithExpiration(expiration))
		require.NoError(t, err)

		encoded, err := invocation.Encode(initial)
		require.NoError(t, err)

		decoded, err := invocation.Decode(encoded)
		require.NoError(t, err)

		require.Equal(t, expiration, *decoded.Expiration())
	})

	t.Run("custom audience", func(t *testing.T) {
		issuer := helpers.RandomSigner(t)
		subject := helpers.RandomDID(t)
		command := helpers.Must(command.Parse("/test/invoke"))(t)
		arguments := invocation.NoArguments
		audience := helpers.RandomDID(t)

		initial, err := invocation.Invoke(issuer, subject, command, arguments, invocation.WithAudience(audience))
		require.NoError(t, err)

		encoded, err := invocation.Encode(initial)
		require.NoError(t, err)

		decoded, err := invocation.Decode(encoded)
		require.NoError(t, err)

		require.Equal(t, &audience, decoded.Audience())
	})

	t.Run("custom auguments", func(t *testing.T) {
		issuer := helpers.RandomSigner(t)
		subject := helpers.RandomDID(t)
		command := helpers.Must(command.Parse("/test/invoke"))(t)
		arguments := helpers.RandomArgs(t)

		initial, err := invocation.Invoke(issuer, subject, command, arguments)
		require.NoError(t, err)

		encoded, err := invocation.Encode(initial)
		require.NoError(t, err)

		decoded, err := invocation.Decode(encoded)
		require.NoError(t, err)

		require.Equal(t, arguments, decoded.Arguments())
	})
}
