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

const (
	Alice = "gCa9UfZv+yI5/rvUIt21DaGI7EZJlzFO1uDc5AyJ30c6/w"
	Bob   = "gCZfj9+RzU2U518TMBNK/fjdGQz34sB4iKE6z+9lQDpCIQ"
	Carol = "gCZC43QGw7ZvYQuKTtBwBy+tdjYrKf0hXU3dd+J0HON5dw"
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

type ErrorModel struct {
	Name string `json:"name"`
}

type InvalidModel struct {
	Name       string       `json:"name"`
	Invocation BytesModel   `json:"invocation"`
	Proofs     []BytesModel `json:"proofs"`
	Error      ErrorModel   `json:"error"`
}

type FixturesModel struct {
	Version    string                `json:"version"`
	Comments   string                `json:"comments"`
	Principals map[string]BytesModel `json:"principals"`
	Valid      []ValidModel          `json:"valid"`
	Invalid    []InvalidModel        `json:"invalid"`
}

var (
	alice = must(ed25519.Decode(must(base64.RawStdEncoding.DecodeString(Alice))))
	bob   = must(ed25519.Decode(must(base64.RawStdEncoding.DecodeString(Bob))))
	carol = must(ed25519.Decode(must(base64.RawStdEncoding.DecodeString(Carol))))
)

var nonce = [][]byte{
	[]byte{1, 2, 3},
	[]byte{4, 5, 6},
	[]byte{7, 8, 9},
}

func main() {
	fixtures := FixturesModel{
		Version:  "1.0.0-rc.1",
		Comments: "Encoded as dag-json. Principals are ed25519 private key bytes with varint(0x1300) prefix.",
		Principals: map[string]BytesModel{
			"alice": BytesModel{Value: alice.Bytes()},
			"bob":   BytesModel{Value: bob.Bytes()},
			"carol": BytesModel{Value: carol.Bytes()},
		},
		Valid: []ValidModel{
			makeValidSelfSignedFixture(),
			makeValidSingleNonTimeBoundedProofFixture(),
			makeValidSingleActiveProofFixture(),
			makeValidMultipleProofsFixture(),
			makeValidMultipleActiveProofsFixture(),
			makeValidPowerlineFixture(),
		},
		Invalid: []InvalidModel{
			makeInvalidMissingProofFixture(),
			makeInvalidExpiredProofFixture(),
			makeInvalidInactiveProofFixture(),
			makeInvalidExpiredInvocationFixture(),
		},
	}

	fmt.Println(string(must(json.MarshalIndent(fixtures, "", "  "))))
}

func makeValidSelfSignedFixture() ValidModel {
	cmd := must(command.Parse("/msg/send"))
	args := datamodel.NewMap()
	inv := must(invocation.Invoke(alice, alice, cmd, args, invocation.WithNoExpiration(), invocation.WithNonce(nonce[0])))

	return ValidModel{
		Name:       "self signed",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{},
	}
}

func makeValidSingleNonTimeBoundedProofFixture() ValidModel {
	cmd := must(command.Parse("/msg/send"))
	dlg0 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithSubject(bob),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[0]),
	))

	inv := must(invocation.Invoke(
		alice,
		bob,
		cmd,
		invocation.NoArguments{},
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link()),
		invocation.WithNonce(nonce[1]),
	))

	return ValidModel{
		Name:       "single non-time bounded proof",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}},
	}
}

func makeValidSingleActiveProofFixture() ValidModel {
	cmd := must(command.Parse("/msg/send"))
	nbf := ucan.UTCUnixTimestamp(must(time.Parse(time.RFC3339, "2025-10-20T11:08:35Z")).Unix())
	dlg0 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithSubject(bob),
		delegation.WithNotBefore(nbf),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[0]),
	))

	inv := must(invocation.Invoke(
		alice,
		bob,
		cmd,
		invocation.NoArguments{},
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link()),
		invocation.WithNonce(nonce[1]),
	))

	return ValidModel{
		Name:       "single active non-expired proof",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}},
	}
}

func makeValidMultipleProofsFixture() ValidModel {
	cmd := must(command.Parse("/msg/send"))
	dlg0 := must(delegation.Delegate(
		carol,
		bob,
		cmd,
		delegation.WithSubject(carol),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[0]),
	))

	dlg1 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithSubject(carol),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[1]),
	))

	inv := must(invocation.Invoke(
		alice,
		carol,
		cmd,
		invocation.NoArguments{},
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link(), dlg1.Link()),
		invocation.WithNonce(nonce[2]),
	))

	return ValidModel{
		Name:       "multiple proofs",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}, {must(delegation.Encode(dlg1))}},
	}
}

func makeValidMultipleActiveProofsFixture() ValidModel {
	cmd := must(command.Parse("/msg/send"))

	dlg0 := must(delegation.Delegate(
		carol,
		bob,
		cmd,
		delegation.WithSubject(carol),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[0]),
	))

	nbf := ucan.UTCUnixTimestamp(must(time.Parse(time.RFC3339, "2025-10-20T11:08:35Z")).Unix())
	dlg1 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithSubject(carol),
		delegation.WithNoExpiration(),
		delegation.WithNotBefore(nbf),
		delegation.WithNonce(nonce[1]),
	))

	inv := must(invocation.Invoke(
		alice,
		carol,
		cmd,
		invocation.NoArguments{},
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link(), dlg1.Link()),
		invocation.WithNonce(nonce[2]),
	))

	return ValidModel{
		Name:       "multiple active proofs",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}, {must(delegation.Encode(dlg1))}},
	}
}

func makeValidPowerlineFixture() ValidModel {
	cmd := must(command.Parse("/msg/send"))

	dlg0 := must(delegation.Delegate(
		carol,
		bob,
		cmd,
		delegation.WithSubject(carol),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[0]),
	))

	dlg1 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithPowerline(true),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[1]),
	))

	inv := must(invocation.Invoke(
		alice,
		carol,
		cmd,
		invocation.NoArguments{},
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link(), dlg1.Link()),
		invocation.WithNonce(nonce[2]),
	))

	return ValidModel{
		Name:       "powerline",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}, {must(delegation.Encode(dlg1))}},
	}
}

func makeInvalidMissingProofFixture() InvalidModel {
	cmd := must(command.Parse("/msg/send"))

	dlg0 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithSubject(bob),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[0]),
	))

	inv := must(invocation.Invoke(
		alice,
		carol,
		cmd,
		invocation.NoArguments{},
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link()),
		invocation.WithNonce(nonce[1]),
	))

	return InvalidModel{
		Name:       "missing proof",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{},
		Error: ErrorModel{
			Name: "UnavailableProof",
		},
	}
}

func makeInvalidExpiredProofFixture() InvalidModel {
	cmd := must(command.Parse("/msg/send"))

	exp := ucan.UTCUnixTimestamp(must(time.Parse(time.RFC3339, "2025-10-20T11:08:35Z")).Unix())
	dlg0 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithSubject(bob),
		delegation.WithExpiration(exp),
		delegation.WithNonce(nonce[0]),
	))

	inv := must(invocation.Invoke(
		alice,
		carol,
		cmd,
		invocation.NoArguments{},
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link()),
		invocation.WithNonce(nonce[1]),
	))

	return InvalidModel{
		Name:       "expired proof",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}},
		Error: ErrorModel{
			Name: "Expired",
		},
	}
}

func makeInvalidInactiveProofFixture() InvalidModel {
	cmd := must(command.Parse("/msg/send"))

	nbf := ucan.UTCUnixTimestamp(must(time.Parse(time.RFC3339, "9999-12-31T23:59:59Z")).Unix())
	dlg0 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithSubject(bob),
		delegation.WithNoExpiration(),
		delegation.WithNotBefore(nbf),
		delegation.WithNonce(nonce[0]),
	))

	inv := must(invocation.Invoke(
		alice,
		carol,
		cmd,
		invocation.NoArguments{},
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link()),
		invocation.WithNonce(nonce[1]),
	))

	return InvalidModel{
		Name:       "inactive proof",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}},
		Error: ErrorModel{
			Name: "TooEarly",
		},
	}
}

func makeInvalidExpiredInvocationFixture() InvalidModel {
	cmd := must(command.Parse("/msg/send"))

	dlg0 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithSubject(bob),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[0]),
	))

	exp := ucan.UTCUnixTimestamp(must(time.Parse(time.RFC3339, "2025-10-20T11:08:35Z")).Unix())
	inv := must(invocation.Invoke(
		alice,
		carol,
		cmd,
		invocation.NoArguments{},
		invocation.WithExpiration(exp),
		invocation.WithProofs(dlg0.Link()),
		invocation.WithNonce(nonce[1]),
	))

	return InvalidModel{
		Name:       "expired invocation",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}},
		Error: ErrorModel{
			Name: "Expired",
		},
	}
}

func must[O any](o O, x error) O {
	if x != nil {
		panic(x)
	}
	return o
}
