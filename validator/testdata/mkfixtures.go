package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/invocation"
)

type BytesModel struct {
	Value []byte
}

func (m BytesModel) MarshalJSON() ([]byte, error) {
	if m.Value == nil {
		return json.Marshal(nil)
	}
	return []byte(fmt.Sprintf(`{"/":{"bytes":"%s"}}`, base64.RawStdEncoding.EncodeToString(m.Value))), nil
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
	Name       string       `json:"name"`
	Invocation BytesModel   `json:"invocation"`
	Proofs     []BytesModel `json:"proofs"`
}

type FixturesModel struct {
	Version    string                `json:"version"`
	Comments   string                `json:"comments"`
	Principals map[string]BytesModel `json:"principals"`
	Valid      []ValidModel          `json:"valid"`
}

const (
	Alice = "gCa9UfZv+yI5/rvUIt21DaGI7EZJlzFO1uDc5AyJ30c6/w"
	Bob   = "gCZfj9+RzU2U518TMBNK/fjdGQz34sB4iKE6z+9lQDpCIQ"
	Carol = "gCZC43QGw7ZvYQuKTtBwBy+tdjYrKf0hXU3dd+J0HON5dw"
)

func main() {
	alice := must(ed25519.Decode(must(base64.RawStdEncoding.DecodeString(Alice))))
	bob := must(ed25519.Decode(must(base64.RawStdEncoding.DecodeString(Bob))))
	carol := must(ed25519.Decode(must(base64.RawStdEncoding.DecodeString(Carol))))

	fixtures := FixturesModel{
		Version:  "1.0.0-rc.1",
		Comments: "Encoded as dag-json. Principals are ed25519 private key bytes with varint(0x1300) prefix.",
		Principals: map[string]BytesModel{
			"alice": BytesModel{Value: alice.Bytes()},
			"bob":   BytesModel{Value: bob.Bytes()},
			"carol": BytesModel{Value: carol.Bytes()},
		},
		Valid: []ValidModel{
			makeValidSelfSignedNoArgsNoProofsNoExpiryFixture(alice, bob, carol),
			makeValidNoPowerlineNoArgsSingleNonTimeBoundedProofNoExpiryFixture(alice, bob, carol),
			makeValidPowerlineNoArgsSingleNonTimeBoundedProofNoExpiryFixture(alice, bob, carol),
			makeValidNoPowerlineNoArgsSingleActiveNonExpiredProofNoExpiryFixture(alice, bob, carol),
			makeValidNoPowerlineNoArgsNonTimeBoundedProofsNoExpiryFixture(alice, bob, carol),
			makeValidNoPowerlineNoArgsActiveNonExpiredProofsNoExpiryFixture(alice, bob, carol),
			makeValidPowerlineNoArgsNonTimeBoundedProofsNoExpiryFixture(alice, bob, carol),
		},
	}

	fmt.Println(string(must(json.MarshalIndent(fixtures, "", "  "))))
}

func makeValidSelfSignedNoArgsNoProofsNoExpiryFixture(alice, bob, carol ucan.Signer) ValidModel {
	cmd := must(command.Parse("/msg/send"))
	args := datamodel.NewMap()
	inv := must(invocation.Invoke(alice, alice, cmd, args, invocation.WithNoExpiration()))

	return ValidModel{
		Name:       "self signed invocation no args, no proofs, no expiry",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{},
	}
}

func makeValidNoPowerlineNoArgsSingleNonTimeBoundedProofNoExpiryFixture(alice, bob, carol ucan.Signer) ValidModel {
	cmd := must(command.Parse("/msg/send"))
	dlg0 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithSubject(bob),
		delegation.WithNoExpiration(),
	))

	args := datamodel.NewMap()
	inv := must(invocation.Invoke(
		alice,
		bob,
		cmd,
		args,
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link()),
	))

	return ValidModel{
		Name:       "invocation with no powerline, no args, single non-time bounded proof, no expiry",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}},
	}
}

func makeValidPowerlineNoArgsSingleNonTimeBoundedProofNoExpiryFixture(alice, bob, carol ucan.Signer) ValidModel {
	cmd := must(command.Parse("/msg/send"))
	dlg0 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithNoExpiration(),
	))

	args := datamodel.NewMap()
	inv := must(invocation.Invoke(
		alice,
		bob,
		cmd,
		args,
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link()),
	))

	return ValidModel{
		Name:       "invocation with powerline, no args, single non-time bounded proof, no expiry",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}},
	}
}

func makeValidNoPowerlineNoArgsSingleActiveNonExpiredProofNoExpiryFixture(alice, bob, carol ucan.Signer) ValidModel {
	cmd := must(command.Parse("/msg/send"))
	nbf := ucan.UTCUnixTimestamp(must(time.Parse(time.RFC3339, "2025-10-20T11:08:35Z")).Unix())
	dlg0 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithSubject(bob),
		delegation.WithNotBefore(nbf),
		delegation.WithNoExpiration(),
	))

	args := datamodel.NewMap()
	inv := must(invocation.Invoke(
		alice,
		bob,
		cmd,
		args,
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link()),
	))

	return ValidModel{
		Name:       "invocation with no powerline, no args, single active non-expired proof, no expiry",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}},
	}
}

func makeValidNoPowerlineNoArgsNonTimeBoundedProofsNoExpiryFixture(alice, bob, carol ucan.Signer) ValidModel {
	cmd := must(command.Parse("/msg/send"))
	dlg0 := must(delegation.Delegate(
		carol,
		bob,
		cmd,
		delegation.WithSubject(carol),
		delegation.WithNoExpiration(),
	))

	dlg1 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithSubject(carol),
		delegation.WithNoExpiration(),
	))

	args := datamodel.NewMap()
	inv := must(invocation.Invoke(
		alice,
		carol,
		cmd,
		args,
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link(), dlg1.Link()),
	))

	return ValidModel{
		Name:       "invocation with no powerline, no args, non non-time bounded proofs, no expiry",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}, {must(delegation.Encode(dlg1))}},
	}
}

func makeValidNoPowerlineNoArgsActiveNonExpiredProofsNoExpiryFixture(alice, bob, carol ucan.Signer) ValidModel {
	cmd := must(command.Parse("/msg/send"))

	dlg0 := must(delegation.Delegate(
		carol,
		bob,
		cmd,
		delegation.WithSubject(carol),
		delegation.WithNoExpiration(),
	))

	nbf := ucan.UTCUnixTimestamp(must(time.Parse(time.RFC3339, "2025-10-20T11:08:35Z")).Unix())
	dlg1 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithSubject(carol),
		delegation.WithNoExpiration(),
		delegation.WithNotBefore(nbf),
	))

	args := datamodel.NewMap()
	inv := must(invocation.Invoke(
		alice,
		carol,
		cmd,
		args,
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link(), dlg1.Link()),
	))

	return ValidModel{
		Name:       "invocation with no powerline, no args, active non-expired proofs, no expiry",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}, {must(delegation.Encode(dlg1))}},
	}
}

func makeValidPowerlineNoArgsNonTimeBoundedProofsNoExpiryFixture(alice, bob, carol ucan.Signer) ValidModel {
	cmd := must(command.Parse("/msg/send"))

	dlg0 := must(delegation.Delegate(
		carol,
		bob,
		cmd,
		delegation.WithNoExpiration(),
	))

	dlg1 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithNoExpiration(),
	))

	args := datamodel.NewMap()
	inv := must(invocation.Invoke(
		alice,
		carol,
		cmd,
		args,
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link(), dlg1.Link()),
	))

	return ValidModel{
		Name:       "invocation with powerline, no args, non-time bounded proofs, no expiry",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}, {must(delegation.Encode(dlg1))}},
	}
}

func must[O any](o O, x error) O {
	if x != nil {
		panic(x)
	}
	return o
}
