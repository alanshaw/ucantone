package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/alanshaw/ucantone/ipld/codec/dagcbor"
	"github.com/alanshaw/ucantone/ipld/datamodel"
	"github.com/alanshaw/ucantone/principal/ed25519"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/command"
	"github.com/alanshaw/ucantone/ucan/delegation"
	ddm "github.com/alanshaw/ucantone/ucan/delegation/datamodel"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/alanshaw/ucantone/ucan/delegation/policy/selector"
	"github.com/alanshaw/ucantone/ucan/invocation"
	idm "github.com/alanshaw/ucantone/ucan/invocation/datamodel"
	"github.com/alanshaw/ucantone/validator"
	"github.com/alanshaw/ucantone/varsig"
	"github.com/alanshaw/ucantone/varsig/common"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	cbg "github.com/whyrusleeping/cbor-gen"
)

// Principals are ed25519 private key bytes with varint(0x1300) prefix.
const (
	Alice = "gCa9UfZv+yI5/rvUIt21DaGI7EZJlzFO1uDc5AyJ30c6/w" // did:key:z6MkgGykN9ARNFjEzowVq4mLP2kL4NsyAaDGXeJFQ5qE1bfg
	Bob   = "gCZfj9+RzU2U518TMBNK/fjdGQz34sB4iKE6z+9lQDpCIQ" // did:key:z6MkmT9j6fVZqzXV8u2wVVSu49gYSRYGSQnduWXF6foAJrqz
	Carol = "gCZC43QGw7ZvYQuKTtBwBy+tdjYrKf0hXU3dd+J0HON5dw" // did:key:z6MkmJceVoQSHs45cReEXoLtWm1wosCG8RLxfKwhxoqzoTkC
	Dave  = "gCY4fdpJOoIaIhEpj4HUj9qfgf8BlW7h3T9IbK9pTddRCw" // did:key:z6Mkh7wJtReCeeT9yDR2nR52omKCayS6zbg8tnW8Jok9CJhk
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
	Version  string         `json:"version"`
	Comments string         `json:"comments"`
	Valid    []ValidModel   `json:"valid"`
	Invalid  []InvalidModel `json:"invalid"`
}

var (
	alice = must(ed25519.Decode(must(base64.RawStdEncoding.DecodeString(Alice))))
	bob   = must(ed25519.Decode(must(base64.RawStdEncoding.DecodeString(Bob))))
	carol = must(ed25519.Decode(must(base64.RawStdEncoding.DecodeString(Carol))))
	dave  = must(ed25519.Decode(must(base64.RawStdEncoding.DecodeString(Dave))))
)

var (
	cmd   = must(command.Parse("/msg/send"))
	nonce = [][]byte{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}
	iat   = ucan.UTCUnixTimestamp(must(time.Parse(time.RFC3339, "2025-10-20T00:00:00Z")).Unix())
)

func main() {
	fixtures := FixturesModel{
		Version:  "1.0.0-rc.1",
		Comments: "Encoded as dag-json.",
		Valid: []ValidModel{
			makeValidSelfSignedFixture(),
			makeValidSingleNonTimeBoundedProofFixture(),
			makeValidSingleActiveProofFixture(),
			makeValidMultipleProofsFixture(),
			makeValidMultipleActiveProofsFixture(),
			makeValidPowerlineFixture(),
			makeValidPolicyMatchFixture(),
		},
		Invalid: []InvalidModel{
			makeInvalidNoProofFixture(),
			makeInvalidMissingProofFixture(),
			makeInvalidExpiredProofFixture(),
			makeInvalidInactiveProofFixture(),
			makeInvalidProofPrincipalAlignmentFixture(),
			makeInvalidInvocationPrincipalAlignmentFixture(),
			makeInvalidProofSubjectAlignmentFixture(),
			makeInvalidInvocationSubjectAlignmentFixture(),
			makeInvalidExpiredInvocationFixture(),
			makeInvalidProofSignatureFixture(),
			makeInvalidInvocationSignatureFixture(),
			makeInvalidPowerlineFixture(),
			makeInvalidPolicyViolationFixture(),
		},
	}

	fmt.Println(string(must(json.MarshalIndent(fixtures, "", "  "))))
}

func makeValidSelfSignedFixture() ValidModel {
	inv := must(invocation.Invoke(
		alice,
		alice,
		cmd,
		invocation.NoArguments{},
		invocation.WithIssuedAt(iat),
		invocation.WithNoExpiration(),
		invocation.WithNonce(nonce[0]),
	))

	return ValidModel{
		Name:       "self signed",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{},
	}
}

func makeValidSingleNonTimeBoundedProofFixture() ValidModel {
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
		invocation.WithIssuedAt(iat),
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
		invocation.WithIssuedAt(iat),
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
		invocation.WithIssuedAt(iat),
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
		invocation.WithIssuedAt(iat),
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
		invocation.WithIssuedAt(iat),
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

func makeValidPolicyMatchFixture() ValidModel {
	dlg0 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithSubject(bob),
		delegation.WithPolicy(
			policy.Equal(must(selector.Parse(".answer")), 42),
		),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[0]),
	))

	inv := must(invocation.Invoke(
		alice,
		bob,
		cmd,
		datamodel.NewMap(datamodel.WithEntry("answer", 42)),
		invocation.WithIssuedAt(iat),
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link()),
		invocation.WithNonce(nonce[1]),
	))

	return ValidModel{
		Name:       "policy match",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}},
	}
}

func makeInvalidNoProofFixture() InvalidModel {
	inv := must(invocation.Invoke(
		alice,
		carol,
		cmd,
		invocation.NoArguments{},
		invocation.WithIssuedAt(iat),
		invocation.WithNoExpiration(),
		invocation.WithNonce(nonce[0]),
	))

	return InvalidModel{
		Name:       "no proof",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{},
		Error:      ErrorModel{validator.InvalidClaimErrorName},
	}
}

func makeInvalidMissingProofFixture() InvalidModel {
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
		invocation.WithIssuedAt(iat),
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link()),
		invocation.WithNonce(nonce[1]),
	))

	return InvalidModel{
		Name:       "missing proof",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{},
		Error:      ErrorModel{validator.UnavailableProofErrorName},
	}
}

func makeInvalidExpiredProofFixture() InvalidModel {
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
		invocation.WithIssuedAt(iat),
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link()),
		invocation.WithNonce(nonce[1]),
	))

	return InvalidModel{
		Name:       "expired proof",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}},
		Error:      ErrorModel{validator.ExpiredErrorName},
	}
}

func makeInvalidInactiveProofFixture() InvalidModel {
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
		invocation.WithIssuedAt(iat),
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link()),
		invocation.WithNonce(nonce[1]),
	))

	return InvalidModel{
		Name:       "inactive proof",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}},
		Error:      ErrorModel{validator.TooEarlyErrorName},
	}
}

func makeInvalidProofPrincipalAlignmentFixture() InvalidModel {
	dlg0 := must(delegation.Delegate(
		dave,
		carol,
		cmd,
		delegation.WithSubject(dave),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[0]),
	))

	dlg1 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithSubject(dave),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[1]),
	))

	inv := must(invocation.Invoke(
		alice,
		carol,
		cmd,
		invocation.NoArguments{},
		invocation.WithIssuedAt(iat),
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link(), dlg1.Link()),
		invocation.WithNonce(nonce[2]),
	))

	return InvalidModel{
		Name:       "proof principal alignment",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}, {must(delegation.Encode(dlg1))}},
		Error:      ErrorModel{validator.PrincipalAlignmentErrorName},
	}
}

func makeInvalidInvocationPrincipalAlignmentFixture() InvalidModel {
	dlg0 := must(delegation.Delegate(
		dave,
		carol,
		cmd,
		delegation.WithSubject(dave),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[0]),
	))

	dlg1 := must(delegation.Delegate(
		carol,
		bob,
		cmd,
		delegation.WithSubject(dave),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[1]),
	))

	inv := must(invocation.Invoke(
		alice,
		dave,
		cmd,
		invocation.NoArguments{},
		invocation.WithIssuedAt(iat),
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link(), dlg1.Link()),
		invocation.WithNonce(nonce[2]),
	))

	return InvalidModel{
		Name:       "invocation principal alignment",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}, {must(delegation.Encode(dlg1))}},
		Error:      ErrorModel{validator.PrincipalAlignmentErrorName},
	}
}

func makeInvalidProofSubjectAlignmentFixture() InvalidModel {
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
		delegation.WithSubject(bob),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[1]),
	))

	inv := must(invocation.Invoke(
		alice,
		carol,
		cmd,
		invocation.NoArguments{},
		invocation.WithIssuedAt(iat),
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link(), dlg1.Link()),
		invocation.WithNonce(nonce[2]),
	))

	return InvalidModel{
		Name:       "proof subject alignment",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}, {must(delegation.Encode(dlg1))}},
		Error:      ErrorModel{validator.SubjectAlignmentErrorName},
	}
}

func makeInvalidInvocationSubjectAlignmentFixture() InvalidModel {
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
		dave,
		cmd,
		invocation.NoArguments{},
		invocation.WithIssuedAt(iat),
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link(), dlg1.Link()),
		invocation.WithNonce(nonce[2]),
	))

	return InvalidModel{
		Name:       "invocation subject alignment",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}, {must(delegation.Encode(dlg1))}},
		Error:      ErrorModel{validator.SubjectAlignmentErrorName},
	}
}

func makeInvalidExpiredInvocationFixture() InvalidModel {
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
		invocation.WithIssuedAt(iat),
		invocation.WithExpiration(exp),
		invocation.WithProofs(dlg0.Link()),
		invocation.WithNonce(nonce[1]),
	))

	return InvalidModel{
		Name:       "expired invocation",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}},
		Error:      ErrorModel{validator.ExpiredErrorName},
	}
}

func makeInvalidProofSignatureFixture() InvalidModel {
	h := must(varsig.Encode(common.Ed25519DagCbor))

	tokenPayload := &ddm.TokenPayloadModel1_0_0_rc1{
		Iss:   bob.DID(),
		Aud:   alice.DID(),
		Sub:   bob.DID(),
		Cmd:   cmd,
		Pol:   ucan.Policy{},
		Nonce: nonce[0],
	}

	sigPayload := ddm.SigPayloadModel{
		Header:                h,
		TokenPayload1_0_0_rc1: tokenPayload,
	}

	model := ddm.EnvelopeModel{
		Signature:  []byte{1, 2, 3},
		SigPayload: sigPayload,
	}

	var dlg0Buf bytes.Buffer
	must0(model.MarshalCBOR(&dlg0Buf))
	dlg0Link := must(cid.V1Builder{
		Codec:  dagcbor.Code,
		MhType: multihash.SHA2_256,
	}.Sum(dlg0Buf.Bytes()))

	inv := must(invocation.Invoke(
		alice,
		carol,
		cmd,
		invocation.NoArguments{},
		invocation.WithIssuedAt(iat),
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0Link),
		invocation.WithNonce(nonce[1]),
	))

	return InvalidModel{
		Name:       "invalid proof signature",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{dlg0Buf.Bytes()}},
		Error:      ErrorModel{validator.InvalidSignatureErrorName},
	}
}

func makeInvalidInvocationSignatureFixture() InvalidModel {
	h := must(varsig.Encode(common.Ed25519DagCbor))

	var args cbg.Deferred
	argsMap := datamodel.NewMap()
	var argsBuf bytes.Buffer
	must0(argsMap.MarshalCBOR(&argsBuf))
	args.Raw = argsBuf.Bytes()

	tokenPayload := &idm.TokenPayloadModel1_0_0_rc1{
		Iss:   alice.DID(),
		Sub:   carol.DID(),
		Cmd:   cmd,
		Args:  args,
		Nonce: nonce[0],
		Iat:   &iat,
	}

	sigPayload := idm.SigPayloadModel{
		Header:                h,
		TokenPayload1_0_0_rc1: tokenPayload,
	}

	model := idm.EnvelopeModel{
		Signature:  []byte{1, 2, 3},
		SigPayload: sigPayload,
	}

	var envBuf bytes.Buffer
	must0(model.MarshalCBOR(&envBuf))

	return InvalidModel{
		Name:       "invalid invocation signature",
		Invocation: BytesModel{envBuf.Bytes()},
		Proofs:     []BytesModel{},
		Error:      ErrorModel{validator.InvalidSignatureErrorName},
	}
}

func makeInvalidPowerlineFixture() InvalidModel {
	dlg0 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithPowerline(true),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[0]),
	))

	inv := must(invocation.Invoke(
		alice,
		carol,
		cmd,
		invocation.NoArguments{},
		invocation.WithIssuedAt(iat),
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link()),
		invocation.WithNonce(nonce[1]),
	))

	return InvalidModel{
		Name:       "invalid powerline",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}},
		Error:      ErrorModel{validator.InvalidClaimErrorName},
	}
}

func makeInvalidPolicyViolationFixture() InvalidModel {
	dlg0 := must(delegation.Delegate(
		bob,
		alice,
		cmd,
		delegation.WithSubject(bob),
		delegation.WithPolicy(
			policy.Equal(must(selector.Parse(".answer")), 42),
		),
		delegation.WithNoExpiration(),
		delegation.WithNonce(nonce[0]),
	))

	inv := must(invocation.Invoke(
		alice,
		bob,
		cmd,
		datamodel.NewMap(datamodel.WithEntry("answer", 41)),
		invocation.WithIssuedAt(iat),
		invocation.WithNoExpiration(),
		invocation.WithProofs(dlg0.Link()),
		invocation.WithNonce(nonce[1]),
	))

	return InvalidModel{
		Name:       "policy violation",
		Invocation: BytesModel{must(invocation.Encode(inv))},
		Proofs:     []BytesModel{{must(delegation.Encode(dlg0))}},
		Error:      ErrorModel{policy.MatchErrorName},
	}
}

func must[O any](o O, x error) O {
	if x != nil {
		panic(x)
	}
	return o
}

func must0(x error) {
	if x != nil {
		panic(x)
	}
}
