package validator_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/alanshaw/ucantone/validator"
	fdm "github.com/alanshaw/ucantone/validator/internal/fixtures/datamodel"
	"github.com/stretchr/testify/require"
)

type NamedError interface {
	error
	Name() string
}

func TestFixtures(t *testing.T) {
	fixturesFile, err := os.Open("./internal/fixtures/invocations.json")
	require.NoError(t, err)

	var fixtures fdm.FixturesModel
	err = fixtures.UnmarshalDagJSON(fixturesFile)
	require.NoError(t, err)

	for _, vector := range fixtures.Valid {
		t.Run("valid "+vector.Name, func(t *testing.T) {
			inv, err := invocation.Decode(vector.Invocation)
			require.NoError(t, err)
			t.Log("invocation", inv.Link())

			proofs := decodeProofs(t, vector.Proofs)
			authority, err := ed25519.Generate()
			require.NoError(t, err)
			vrf := authority.Verifier()

			// TODO: capability details in the vector?
			cmd, err := command.Parse("/msg/send")
			require.NoError(t, err)

			opts := []validator.Option{
				validator.WithProofResolver(newMapProofResolver(proofs)),
			}
			cap := validator.NewCapability[invocation.UnknownArguments](cmd, ucan.Policy{})
			authorization, err := validator.Access(t.Context(), vrf, cap, inv, opts...)
			require.NoError(t, err, "validation should have passed for invocation with %s", vector.Description)

			_, err = authorization.Task.BindArguments()
			require.NoError(t, err)
		})
	}

	for _, vector := range fixtures.Invalid {
		t.Run("invalid "+vector.Name, func(t *testing.T) {
			inv, err := invocation.Decode(vector.Invocation)
			require.NoError(t, err)
			t.Log("invocation", inv.Link())

			proofs := decodeProofs(t, vector.Proofs)
			authority, err := ed25519.Generate()
			require.NoError(t, err)
			vrf := authority.Verifier()

			// TODO: capability details in the vector?
			cmd, err := command.Parse("/msg/send")
			require.NoError(t, err)

			opts := []validator.Option{
				validator.WithProofResolver(newMapProofResolver(proofs)),
			}
			cap := validator.NewCapability[invocation.UnknownArguments](cmd, ucan.Policy{})
			_, err = validator.Access(t.Context(), vrf, cap, inv, opts...)
			require.Error(t, err, "validation should not have passed for invocation because %s", vector.Description)
			t.Log(err)

			var namedErr NamedError
			require.True(t, errors.As(err, &namedErr))
			require.Equal(t, vector.Error.Name, namedErr.Name())
		})
	}
}

func newMapProofResolver(proofs map[ucan.Link]ucan.Delegation) validator.ProofResolverFunc {
	return func(_ context.Context, link ucan.Link) (ucan.Delegation, error) {
		dlg, ok := proofs[link]
		if !ok {
			return nil, validator.NewUnavailableProofError(link, errors.New("not provided"))
		}
		return dlg, nil
	}
}

func decodeProofs(t *testing.T, vectorProofs [][]byte) map[ucan.Link]ucan.Delegation {
	proofs := map[ucan.Link]ucan.Delegation{}
	for _, p := range vectorProofs {
		dlg, err := delegation.Decode(p)
		require.NoError(t, err)
		proofs[dlg.Link()] = dlg
		t.Log("proof", dlg.Link())
	}
	return proofs
}
