package did_test

import (
	"encoding/json"
	"testing"

	"github.com/fil-forge/ucantone/did"
	"github.com/stretchr/testify/require"
)

func TestVerificationRelationship_Add(t *testing.T) {
	d, err := did.Parse("did:example:123456789abcdefghi")
	require.NoError(t, err)
	doc := did.NewDocument(d)
	vm := did.VerificationMethod{
		ID:         doc.Fragment("key-1"),
		Controller: d,
		Type:       did.MultikeyVerificationMethodType,
		Material:   did.GenericMap{did.MultikeyPublicKeyMultibaseProp: "zABC"},
	}
	err = doc.VerificationMethods.Add(vm)
	require.NoError(t, err)

	require.Equal(t, 0, doc.Authentication.Len())

	err = doc.Authentication.Add(vm)
	require.NoError(t, err)
	require.Equal(t, 1, doc.Authentication.Len())

	var authVMIds []string
	for _, authVM := range doc.Authentication.All() {
		authVMIds = append(authVMIds, authVM.ID.String())
	}
	require.Equal(t, []string{vm.ID.String()}, authVMIds)
}

// plcShapedDoc mirrors what the PLC directory serves: verification methods
// but NO relationship sections. Relationship sections default to endorsing
// every method (Document.UnmarshalJSON), and All() must honor that —
// regression: All() only consulted the explicit references, so every
// relationship of a PLC-resolved document verified against zero keys and
// every did:plc-issued token failed signature verification.
const plcShapedDoc = `{
	"@context": [
		"https://www.w3.org/ns/did/v1",
		"https://w3id.org/security/multikey/v1"
	],
	"id": "did:plc:li7co67qxxjdhfwum4ivdlay",
	"alsoKnownAs": [],
	"verificationMethod": [
		{
			"id": "did:plc:li7co67qxxjdhfwum4ivdlay#hilt",
			"type": "Multikey",
			"controller": "did:plc:li7co67qxxjdhfwum4ivdlay",
			"publicKeyMultibase": "zQ3shY3SRmBHD2k3MbUG2k9TyVxtSVFEAGPCScnkBnKya18qj"
		}
	],
	"service": []
}`

func TestVerificationRelationship_UndeclaredDefaultsToAllMethods(t *testing.T) {
	var doc did.Document
	require.NoError(t, doc.UnmarshalJSON([]byte(plcShapedDoc)))

	for name, rel := range map[string]*did.VerificationRelationship{
		"authentication":       doc.Authentication,
		"assertionMethod":      doc.AssertionMethod,
		"keyAgreement":         doc.KeyAgreement,
		"capabilityInvocation": doc.CapabilityInvocation,
		"capabilityDelegation": doc.CapabilityDelegation,
	} {
		vms := rel.All()
		require.Len(t, vms, 1, "undeclared %s must endorse all document methods", name)
		require.Equal(t, "did:plc:li7co67qxxjdhfwum4ivdlay#hilt", vms[0].ID.String())
	}
}

func TestVerificationRelationship_DeclaredEmptyYieldsNoMethods(t *testing.T) {
	docJSON := `{
		"@context": "https://www.w3.org/ns/did/v1",
		"id": "did:example:123",
		"verificationMethod": [
			{
				"id": "did:example:123#key-1",
				"type": "Multikey",
				"controller": "did:example:123",
				"publicKeyMultibase": "zABC"
			}
		],
		"capabilityDelegation": []
	}`
	var doc did.Document
	require.NoError(t, doc.UnmarshalJSON([]byte(docJSON)))

	// Explicitly declared empty: exactly zero methods.
	require.Empty(t, doc.CapabilityDelegation.All())
	// Undeclared sibling still defaults to all methods.
	require.Len(t, doc.CapabilityInvocation.All(), 1)
}

func TestVerificationRelationship_NullIsUndeclared(t *testing.T) {
	docJSON := `{
		"@context": "https://www.w3.org/ns/did/v1",
		"id": "did:example:123",
		"verificationMethod": [
			{
				"id": "did:example:123#key-1",
				"type": "Multikey",
				"controller": "did:example:123",
				"publicKeyMultibase": "zABC"
			}
		],
		"capabilityDelegation": null
	}`
	var doc did.Document
	require.NoError(t, doc.UnmarshalJSON([]byte(docJSON)))

	// A null relationship declares nothing, so it defaults to all methods —
	// unlike an explicitly empty array, which endorses none.
	require.Len(t, doc.CapabilityDelegation.All(), 1)
	require.True(t, doc.CapabilityDelegation.IsZero())
}

// TestVerificationRelationship_MarshalRoundTrip guards the declared/undeclared
// distinction across serialization: a declared empty relationship endorses no
// methods, so dropping it on marshal would silently widen it to endorsing every
// method in the document.
func TestVerificationRelationship_MarshalRoundTrip(t *testing.T) {
	docJSON := `{
		"@context": "https://www.w3.org/ns/did/v1",
		"id": "did:example:123",
		"verificationMethod": [
			{
				"id": "did:example:123#key-1",
				"type": "Multikey",
				"controller": "did:example:123",
				"publicKeyMultibase": "zABC"
			}
		],
		"capabilityDelegation": []
	}`
	var doc did.Document
	require.NoError(t, doc.UnmarshalJSON([]byte(docJSON)))

	b, err := json.Marshal(doc)
	require.NoError(t, err)

	var fields map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(b, &fields))
	// The declared relationship is emitted as an empty array...
	require.Equal(t, "[]", string(fields["capabilityDelegation"]))
	// ...and the undeclared ones stay absent, rather than being emitted empty
	// and thereby restricted to nothing.
	for _, name := range []string{
		"authentication", "assertionMethod", "keyAgreement", "capabilityInvocation",
	} {
		require.NotContains(t, fields, name)
	}

	var reparsed did.Document
	require.NoError(t, json.Unmarshal(b, &reparsed))
	require.Empty(t, reparsed.CapabilityDelegation.All(), "declared empty must survive a round trip")
	require.Len(t, reparsed.CapabilityInvocation.All(), 1, "undeclared must still default to all methods")
}

// TestVerificationRelationship_StandaloneMarshal guards marshaling a
// relationship outside a Document, where no omitzero drops undeclared ones:
// undeclared must emit null (which unmarshals back as undeclared), not "[]",
// which would round-trip as declared-empty and narrow "endorses everything"
// to "endorses nothing".
func TestVerificationRelationship_StandaloneMarshal(t *testing.T) {
	var undeclared did.VerificationRelationship
	b, err := json.Marshal(&undeclared)
	require.NoError(t, err)
	require.Equal(t, "null", string(b))

	var reparsed did.VerificationRelationship
	require.NoError(t, json.Unmarshal(b, &reparsed))
	require.True(t, reparsed.IsZero(), "undeclared must survive a standalone round trip")
}

// TestVerificationRelationship_UnmarshalReuse guards against unmarshaling
// into an already-populated instance accumulating references from the
// previous document.
func TestVerificationRelationship_UnmarshalReuse(t *testing.T) {
	var vr did.VerificationRelationship
	require.NoError(t, json.Unmarshal([]byte(`["did:example:123#key-1"]`), &vr))
	require.NoError(t, json.Unmarshal([]byte(`["did:example:123#key-2"]`), &vr))

	require.Equal(t, 1, vr.Len(), "reuse must replace references, not accumulate them")
	require.Equal(t, "did:example:123#key-2", vr.Get(0).String())

	// Per json.Unmarshaler convention null is a no-op — it leaves existing
	// state untouched, like json.Unmarshal on a slice.
	require.NoError(t, json.Unmarshal([]byte(`null`), &vr))
	require.Equal(t, 1, vr.Len())
	require.False(t, vr.IsZero())
}

func TestVerificationRelationship_DeclaredSubset(t *testing.T) {
	docJSON := `{
		"@context": "https://www.w3.org/ns/did/v1",
		"id": "did:example:123",
		"verificationMethod": [
			{
				"id": "did:example:123#key-1",
				"type": "Multikey",
				"controller": "did:example:123",
				"publicKeyMultibase": "zABC"
			},
			{
				"id": "did:example:123#key-2",
				"type": "Multikey",
				"controller": "did:example:123",
				"publicKeyMultibase": "zDEF"
			}
		],
		"capabilityDelegation": ["did:example:123#key-2"]
	}`
	var doc did.Document
	require.NoError(t, doc.UnmarshalJSON([]byte(docJSON)))

	vms := doc.CapabilityDelegation.All()
	require.Len(t, vms, 1)
	require.Equal(t, "did:example:123#key-2", vms[0].ID.String())
	// Undeclared sibling sees both.
	require.Len(t, doc.CapabilityInvocation.All(), 2)
}
