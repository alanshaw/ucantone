package validator_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/testing/helpers"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/alanshaw/ucantone/validator"
	"github.com/stretchr/testify/require"
)

type BytesModel struct {
	Value []byte
}

func (m *BytesModel) UnmarshalJSON(data []byte) error {
	var outer map[string]map[string]string
	err := json.Unmarshal(data, &outer)
	if err != nil {
		return err
	}
	if outer == nil {
		return nil
	}
	if len(outer) != 1 {
		return errors.New("invalid IPLD value: extraneous fields")
	}
	inner, ok := outer["/"]
	if !ok {
		return errors.New(`invalid IPLD value: missing field: "/"`)
	}
	if len(inner) != 1 {
		return errors.New("invalid IPLD value: extraneous fields")
	}
	b64bytes, ok := inner["bytes"]
	if !ok {
		return errors.New(`invalid IPLD value: missing field: "bytes"`)
	}
	bytes, err := base64.RawStdEncoding.DecodeString(b64bytes)
	if err != nil {
		return fmt.Errorf("decoding base64 bytes: %w", err)
	}
	m.Value = bytes
	return nil
}

type ValidModel struct {
	Name       string
	Invocation BytesModel
	Proofs     []BytesModel
}

type ErrorModel struct {
	Name string
}

type InvalidModel struct {
	Name       string
	Invocation BytesModel
	Proofs     []BytesModel
	Error      ErrorModel
}

type FixturesModel struct {
	Version    string
	Comments   string
	Principals map[string]BytesModel
	Valid      []ValidModel
	Invalid    []InvalidModel
}

func TestFixtures(t *testing.T) {
	fixtureBytes := helpers.Must(os.ReadFile("./testdata/fixtures/executables.json"))(t)

	var fixtures FixturesModel
	err := json.Unmarshal(fixtureBytes, &fixtures)
	require.NoError(t, err)

	principals := map[string]ucan.Signer{}
	for name, bytes := range fixtures.Principals {
		signer := helpers.Must(ed25519.Decode(bytes.Value))(t)
		principals[signer.DID().String()] = signer
		t.Logf("%s: %s", name, signer.DID())
	}

	for _, vector := range fixtures.Valid {
		t.Run(vector.Name, func(t *testing.T) {
			inv, err := invocation.Decode(vector.Invocation.Value)
			require.NoError(t, err)
			t.Log("invocation", inv.Link())

			proofs := map[ucan.Link]ucan.Delegation{}
			for _, p := range vector.Proofs {
				dlg, err := delegation.Decode(p.Value)
				require.NoError(t, err)
				proofs[dlg.Link()] = dlg
				t.Log("proof", dlg.Link())
			}

			resolveProof := func(_ context.Context, link ucan.Link) (ucan.Delegation, error) {
				dlg, ok := proofs[link]
				if !ok {
					return nil, validator.NewUnavailableProofError(link, errors.New("not provided"))
				}
				return dlg, nil
			}

			authority, err := ed25519.Generate()
			require.NoError(t, err)

			// TODO: capability details in the vector?
			cmd, err := command.Parse("/msg/send")
			require.NoError(t, err)

			cap := validator.NewCapability[invocation.NoArguments](cmd, ucan.Policy{})
			_, err = validator.Access(
				t.Context(),
				authority.Verifier(),
				cap,
				inv,
				validator.WithProofResolver(resolveProof),
			)
			require.NoError(t, err)
		})
	}
}
