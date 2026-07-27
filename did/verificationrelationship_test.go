package did_test

import (
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
