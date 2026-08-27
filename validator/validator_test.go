package validator_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multibase"
	"github.com/multiformats/go-varint"
	"github.com/stretchr/testify/require"

	"github.com/fil-forge/ucantone/absentee"
	"github.com/fil-forge/ucantone/did"
	"github.com/fil-forge/ucantone/did/key"
	"github.com/fil-forge/ucantone/ipld/datamodel"
	"github.com/fil-forge/ucantone/testutil"
	"github.com/fil-forge/ucantone/ucan"
	"github.com/fil-forge/ucantone/ucan/command"
	"github.com/fil-forge/ucantone/ucan/container"
	"github.com/fil-forge/ucantone/ucan/delegation"
	"github.com/fil-forge/ucantone/ucan/delegation/policy"
	"github.com/fil-forge/ucantone/ucan/invocation"
	"github.com/fil-forge/ucantone/validator"
	verrs "github.com/fil-forge/ucantone/validator/errors"
	fdm "github.com/fil-forge/ucantone/validator/internal/fixtures/datamodel"
)

const (
	past   ucan.UnixTimestamp = 1000000000 // 2001-09-09
	future ucan.UnixTimestamp = 9999999999 // 2286-11-20
	now    ucan.UnixTimestamp = 1746748800 // 2025-05-09 (fixed validation time for tests)
)

// badIssuer is a Signer that produces invalid signatures, for testing purposes.
type badIssuer struct{ ucan.Issuer }

func (b badIssuer) Sign(msg []byte) []byte {
	sig := b.Issuer.Sign(msg)
	sig[0] ^= 0xff // flip a bit
	return sig
}

func TestValidate(t *testing.T) {
	crankWidget := testutil.Must(command.Parse("/widget/crank"))(t)

	t.Run("validates with root authority", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		inv, err := invocation.Invoke(subject, subject.DID(), crankWidget, datamodel.Map{})
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv)
		require.NoError(t, err)
	})

	t.Run("rejects with a bad signature", func(t *testing.T) {
		subject := badIssuer{testutil.RandomIssuer(t)}
		inv, err := invocation.Invoke(subject, subject.DID(), crankWidget, datamodel.Map{})
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv)
		require.Error(t, err)
	})

	t.Run("rejects with unauthorized invoker", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		invoker := testutil.RandomIssuer(t)

		inv, err := invocation.Invoke(subject, invoker.DID(), crankWidget, datamodel.Map{})
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv)
		require.Error(t, err)
	})

	t.Run("validates with subject → invoker", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		invoker := testutil.RandomIssuer(t)

		del, err := delegation.Delegate(subject, invoker.DID(), subject.DID(), crankWidget)
		require.NoError(t, err)

		inv, err := invocation.Invoke(
			invoker,
			subject.DID(),
			crankWidget,
			datamodel.Map{},
			invocation.WithProofs(del.Link()),
		)
		require.NoError(t, err)

		err = validator.ValidateInvocation(
			t.Context(),
			inv,
			validator.WithProofResolver(
				validator.ProofsFromContainer(
					container.New(container.WithDelegations(del)),
				),
			),
		)
		require.NoError(t, err)
	})

	t.Run("rejects an expired invocation", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		inv, err := invocation.Invoke(subject, subject.DID(), crankWidget, datamodel.Map{},
			invocation.WithExpiration(past),
		)
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv, validator.WithValidationTime(now))
		require.Error(t, err)
	})

	t.Run("accepts an invocation with a future expiry", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		inv, err := invocation.Invoke(subject, subject.DID(), crankWidget, datamodel.Map{},
			invocation.WithExpiration(future),
		)
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv, validator.WithValidationTime(now))
		require.NoError(t, err)
	})

	t.Run("rejects a proof that is not yet valid", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		invoker := testutil.RandomIssuer(t)

		del, err := delegation.Delegate(subject, invoker.DID(), subject.DID(), crankWidget,
			delegation.WithNotBefore(future),
		)
		require.NoError(t, err)

		inv, err := invocation.Invoke(
			invoker,
			subject.DID(),
			crankWidget,
			datamodel.Map{},
			invocation.WithProofs(del.Link()),
		)
		require.NoError(t, err)

		err = validator.ValidateInvocation(
			t.Context(),
			inv,
			validator.WithValidationTime(now),
			validator.WithProofResolver(
				validator.ProofsFromContainer(
					container.New(container.WithDelegations(del)),
				),
			),
		)
		require.Error(t, err)
	})

	t.Run("rejects when final proof audience does not match invoker", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		invoker := testutil.RandomIssuer(t)
		other := testutil.RandomIssuer(t)

		// Delegation goes to other, but invoker invokes
		del, err := delegation.Delegate(subject, other.DID(), subject.DID(), crankWidget)
		require.NoError(t, err)

		inv, err := invocation.Invoke(
			invoker,
			subject.DID(),
			crankWidget,
			datamodel.Map{},
			invocation.WithProofs(del.Link()),
		)
		require.NoError(t, err)

		err = validator.ValidateInvocation(
			t.Context(),
			inv,
			validator.WithProofResolver(
				validator.ProofsFromContainer(
					container.New(container.WithDelegations(del)),
				),
			),
		)
		require.Error(t, err)
	})

	t.Run("rejects a broken mid-chain (issuer mismatch)", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		alice := testutil.RandomIssuer(t)
		bob := testutil.RandomIssuer(t)
		eve := testutil.RandomIssuer(t)

		del1, err := delegation.Delegate(subject, alice.DID(), subject.DID(), crankWidget)
		require.NoError(t, err)
		// del2 is from eve, not alice — breaks the chain
		del2, err := delegation.Delegate(eve, bob.DID(), subject.DID(), crankWidget)
		require.NoError(t, err)

		inv, err := invocation.Invoke(
			bob,
			subject.DID(),
			crankWidget,
			datamodel.Map{},
			invocation.WithProofs(del1.Link(), del2.Link()),
		)
		require.NoError(t, err)

		err = validator.ValidateInvocation(
			t.Context(),
			inv,
			validator.WithProofResolver(
				validator.ProofsFromContainer(
					container.New(container.WithDelegations(del1, del2)),
				),
			),
		)
		require.Error(t, err)
	})

	t.Run("validates with subject → alice → bob", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		alice := testutil.RandomIssuer(t)
		bob := testutil.RandomIssuer(t)

		del1, err := delegation.Delegate(subject, alice.DID(), subject.DID(), crankWidget)
		require.NoError(t, err)
		del2, err := delegation.Delegate(alice, bob.DID(), subject.DID(), crankWidget)
		require.NoError(t, err)

		inv, err := invocation.Invoke(
			bob,
			subject.DID(),
			crankWidget,
			datamodel.Map{},
			invocation.WithProofs(del1.Link(), del2.Link()),
		)
		require.NoError(t, err)

		err = validator.ValidateInvocation(
			t.Context(),
			inv,
			validator.WithProofResolver(
				validator.ProofsFromContainer(
					container.New(container.WithDelegations(del1, del2)),
				),
			),
		)
		require.NoError(t, err)
	})

	t.Run("rejects when a referenced proof cannot be resolved", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		invoker := testutil.RandomIssuer(t)

		del, err := delegation.Delegate(subject, invoker.DID(), subject.DID(), crankWidget)
		require.NoError(t, err)

		inv, err := invocation.Invoke(
			invoker,
			subject.DID(),
			crankWidget,
			datamodel.Map{},
			invocation.WithProofs(del.Link()),
		)
		require.NoError(t, err)

		// No WithProofResolver — default ProofUnavailable fires
		err = validator.ValidateInvocation(t.Context(), inv)
		require.Error(t, err)
	})

	// https://github.com/ucan-wg/delegation#powerline
	t.Run("validates with powerline delegation in chain", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		alice := testutil.RandomIssuer(t)
		bob := testutil.RandomIssuer(t)

		del1, err := delegation.Delegate(subject, alice.DID(), subject.DID(), crankWidget)
		require.NoError(t, err)
		// del2 is a powerline delegation, where alice delegates `/widget/crank` to
		// bob for any subject she herself is authorized to `/widget/crank`.
		del2, err := delegation.Delegate(alice, bob.DID(), did.Undef, crankWidget)
		require.NoError(t, err)

		inv, err := invocation.Invoke(
			bob,
			subject.DID(),
			crankWidget,
			datamodel.Map{},
			invocation.WithProofs(del1.Link(), del2.Link()),
		)
		require.NoError(t, err)

		err = validator.ValidateInvocation(
			t.Context(),
			inv,
			validator.WithProofResolver(
				validator.ProofsFromContainer(
					container.New(container.WithDelegations(del1, del2)),
				),
			),
		)
		require.NoError(t, err)
	})

	// Explicitly disallowed by spec:
	// https://github.com/ucan-wg/delegation#powerline
	t.Run("rejects a powerline delegation at root of chain", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		invoker := testutil.RandomIssuer(t)

		// Root delegation has nil subject — invalid per spec.
		del, err := delegation.Delegate(subject, invoker.DID(), did.Undef, crankWidget)
		require.NoError(t, err)

		inv, err := invocation.Invoke(
			invoker,
			subject.DID(),
			crankWidget,
			datamodel.Map{},
			invocation.WithProofs(del.Link()),
		)
		require.NoError(t, err)

		err = validator.ValidateInvocation(
			t.Context(),
			inv,
			validator.WithProofResolver(
				validator.ProofsFromContainer(
					container.New(container.WithDelegations(del)),
				),
			),
		)
		require.Error(t, err)
	})

	t.Run("accepts a proof with a NotBefore in the past", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		invoker := testutil.RandomIssuer(t)

		del, err := delegation.Delegate(subject, invoker.DID(), subject.DID(), crankWidget,
			delegation.WithNotBefore(past),
		)
		require.NoError(t, err)

		inv, err := invocation.Invoke(
			invoker,
			subject.DID(),
			crankWidget,
			datamodel.Map{},
			invocation.WithProofs(del.Link()),
		)
		require.NoError(t, err)

		err = validator.ValidateInvocation(
			t.Context(),
			inv,
			validator.WithValidationTime(now),
			validator.WithProofResolver(
				validator.ProofsFromContainer(
					container.New(container.WithDelegations(del)),
				),
			),
		)
		require.NoError(t, err)
	})

	t.Run("with a policy on a delegation", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		invoker := testutil.RandomIssuer(t)

		del, err := delegation.Delegate(subject, invoker.DID(), subject.DID(), crankWidget,
			delegation.WithPolicyBuilder(policy.Equal(".answer", 42)),
		)
		require.NoError(t, err)

		resolveProof := validator.ProofsFromContainer(
			container.New(container.WithDelegations(del)),
		)

		t.Run("accepts an invocation whose arguments satisfy the policy", func(t *testing.T) {
			inv, err := invocation.Invoke(
				invoker,
				subject.DID(),
				crankWidget,
				datamodel.Map{"answer": 42},
				invocation.WithProofs(del.Link()),
			)
			require.NoError(t, err)

			err = validator.ValidateInvocation(
				t.Context(),
				inv,
				validator.WithProofResolver(resolveProof),
			)
			require.NoError(t, err)
		})

		t.Run("rejects an invocation whose arguments violate the policy", func(t *testing.T) {
			inv, err := invocation.Invoke(
				invoker,
				subject.DID(),
				crankWidget,
				datamodel.Map{"answer": 41},
				invocation.WithProofs(del.Link()),
			)
			require.NoError(t, err)

			err = validator.ValidateInvocation(
				t.Context(),
				inv,
				validator.WithProofResolver(resolveProof),
			)
			require.Error(t, err)
		})
	})

	t.Run("rejects with incorrect subject in chain", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		alice := testutil.RandomIssuer(t)
		bob := testutil.RandomIssuer(t)
		unrelatedSubject := testutil.RandomIssuer(t)

		del1, err := delegation.Delegate(subject, alice.DID(), subject.DID(), crankWidget)
		require.NoError(t, err)
		// del2 is about the wrong subject.
		del2, err := delegation.Delegate(alice, bob.DID(), unrelatedSubject.DID(), crankWidget)
		require.NoError(t, err)

		inv, err := invocation.Invoke(
			bob,
			subject.DID(),
			crankWidget,
			datamodel.Map{},
			invocation.WithProofs(del1.Link(), del2.Link()),
		)
		require.NoError(t, err)

		err = validator.ValidateInvocation(
			t.Context(),
			inv,
			validator.WithProofResolver(
				validator.ProofsFromContainer(
					container.New(container.WithDelegations(del1, del2)),
				),
			),
		)
		require.Error(t, err)
	})

	t.Run("rejects when the signing key is expired", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)

		// Build a DID document with an expired Multikey VM.
		expires := did.DateTimeStamp(time.Unix(int64(past), 0))
		resolver := expiredKeyResolver(t, &expires, nil)

		inv, err := invocation.Invoke(subject, subject.DID(), crankWidget, datamodel.Map{})
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv,
			validator.WithDIDResolver(resolver),
			validator.WithValidationTime(now),
		)
		require.ErrorContains(t, err, "expired")
	})

	t.Run("rejects when the signing key is revoked", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)

		revoked := did.DateTimeStamp(time.Unix(int64(past), 0))
		resolver := expiredKeyResolver(t, nil, &revoked)

		inv, err := invocation.Invoke(subject, subject.DID(), crankWidget, datamodel.Map{})
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv,
			validator.WithDIDResolver(resolver),
			validator.WithValidationTime(now),
		)
		require.ErrorContains(t, err, "revoked")
	})

	t.Run("accepts when the signing key has not yet expired", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)

		expires := did.DateTimeStamp(time.Unix(int64(future), 0))
		resolver := expiredKeyResolver(t, &expires, nil)

		inv, err := invocation.Invoke(subject, subject.DID(), crankWidget, datamodel.Map{})
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv,
			validator.WithDIDResolver(resolver),
			validator.WithValidationTime(now),
		)
		require.NoError(t, err)
	})

	t.Run("with non-standard signature in chain", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		alice := absentee.From(testutil.Must(did.Parse("did:example:alice"))(t))
		bob := testutil.RandomIssuer(t)

		del1, err := delegation.Delegate(subject, alice.DID(), subject.DID(), crankWidget)
		require.NoError(t, err)
		// del2 is "signed" by alice, who is an absentee signer and produces a
		// non-standard signature.
		del2, err := delegation.Delegate(alice, bob.DID(), did.Undef, crankWidget)
		require.NoError(t, err)

		inv, err := invocation.Invoke(
			bob,
			subject.DID(),
			crankWidget,
			datamodel.Map{},
			invocation.WithProofs(del1.Link(), del2.Link()),
		)
		require.NoError(t, err)

		resolveProof := validator.ProofsFromContainer(
			container.New(container.WithDelegations(del1, del2)),
		)

		t.Run("rejects by default", func(t *testing.T) {
			err = validator.ValidateInvocation(
				t.Context(),
				inv,
				validator.WithProofResolver(resolveProof),
				validator.WithDIDResolver(did.ResolverMap{
					"key": key.Resolver,
					"example": did.ResolverFunc(func(ctx context.Context, d did.DID) (did.Document, error) {
						require.Fail(t, "shouldn't try to resolve a verifier for a non-standard signature")
						return did.Document{}, nil
					}),
				}),
			)
			require.ErrorContains(t, err, "no non-standard signature verifier configured")
		})

		t.Run("rejects according to non-standard signature verifier", func(t *testing.T) {
			err = validator.ValidateInvocation(
				t.Context(),
				inv,
				validator.WithProofResolver(resolveProof),
				validator.WithNonStandardSignatureVerifier(
					func(ctx context.Context, token ucan.Token, meta ucan.Container) error {
						require.Equal(t, del2.Link(), token.Link(), "should be asked to verify the non-standard signature for the correct token")
						return errors.New("non-standard error failed as expected")
					},
				),
			)
			require.ErrorContains(t, err, "non-standard error failed as expected")
		})

		t.Run("validates according to non-standard signature verifier", func(t *testing.T) {
			err = validator.ValidateInvocation(
				t.Context(),
				inv,
				validator.WithProofResolver(resolveProof),
				validator.WithNonStandardSignatureVerifier(
					func(ctx context.Context, token ucan.Token, meta ucan.Container) error {
						require.Equal(t, del2.Link(), token.Link(), "should be asked to verify the non-standard signature for the correct token")
						return nil
					},
				),
			)
			require.NoError(t, err)
		})
	})
}

// TestRelationshipFallback covers the optional capability relationships (DID
// core §5.3): a document that expresses no capabilityInvocation /
// capabilityDelegation authorizes all of its verification methods, while a
// document that expresses one restricts verification to the methods listed —
// including an explicitly empty relationship, which authorizes nothing.
func TestRelationshipFallback(t *testing.T) {
	crankWidget := testutil.Must(command.Parse("/widget/crank"))(t)

	// plcShapedResolver serves a document parsed from JSON that carries only
	// verificationMethod — the shape did:plc directories emit.
	plcShapedResolver := did.ResolverFunc(func(_ context.Context, d did.DID) (did.Document, error) {
		docJSON := fmt.Sprintf(`{
			"id": %q,
			"verificationMethod": [{
				"id": "%s#key",
				"type": %q,
				"controller": %q,
				"publicKeyMultibase": %q
			}]
		}`, d, d, did.MultikeyVerificationMethodType, d, d.Identifier())
		var doc did.Document
		if err := json.Unmarshal([]byte(docJSON), &doc); err != nil {
			return did.Document{}, err
		}
		return doc, nil
	})

	t.Run("invocation verifies against a document without relationships", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)

		inv, err := invocation.Invoke(subject, subject.DID(), crankWidget, datamodel.Map{})
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv,
			validator.WithDIDResolver(plcShapedResolver))
		require.NoError(t, err)
	})

	t.Run("delegation in chain verifies against a document without relationships", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		bob := testutil.RandomIssuer(t)

		del, err := delegation.Delegate(subject, bob.DID(), subject.DID(), crankWidget)
		require.NoError(t, err)
		inv, err := invocation.Invoke(bob, subject.DID(), crankWidget, datamodel.Map{},
			invocation.WithProofs(del.Link()))
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv,
			validator.WithDIDResolver(plcShapedResolver),
			validator.WithProofResolver(validator.ProofsFromContainer(
				container.New(container.WithDelegations(del)))),
		)
		require.NoError(t, err)
	})

	t.Run("populated relationship still restricts to its listed methods", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		other := testutil.RandomIssuer(t) // a key the subject does NOT sign with

		resolver := did.ResolverFunc(func(_ context.Context, d did.DID) (did.Document, error) {
			doc := did.NewDocument(d)
			signerVM := did.VerificationMethod{
				ID:         doc.Fragment("signer"),
				Controller: d,
				Type:       did.MultikeyVerificationMethodType,
				Material:   did.GenericMap{did.MultikeyPublicKeyMultibaseProp: d.Identifier()},
			}
			otherVM := did.VerificationMethod{
				ID:         doc.Fragment("other"),
				Controller: d,
				Type:       did.MultikeyVerificationMethodType,
				Material:   did.GenericMap{did.MultikeyPublicKeyMultibaseProp: other.DID().Identifier()},
			}
			if err := doc.VerificationMethods.Add(signerVM); err != nil {
				return did.Document{}, err
			}
			// Only the non-signing key is authorized for capability invocation,
			// so verification must NOT fall back to the signer's method.
			if err := doc.CapabilityInvocation.Add(otherVM); err != nil {
				return did.Document{}, err
			}
			return doc, nil
		})

		inv, err := invocation.Invoke(subject, subject.DID(), crankWidget, datamodel.Map{})
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv,
			validator.WithDIDResolver(resolver))
		require.Error(t, err)
	})

	// undecodableKeyMultibase is a well-formed Multikey string whose multicodec
	// key type has no verifier registered: x25519-pub (0xec), a key-agreement
	// key that can never sign. did:plc documents publish such keys next to the
	// signing key.
	undecodableKeyMultibase := func(t *testing.T) string {
		bytes := append(varint.ToUvarint(0xec), make([]byte, 32)...)
		mb, err := multibase.Encode(multibase.Base58BTC, bytes)
		require.NoError(t, err)
		return mb
	}

	// method pairs a verification method fragment with its Multikey material.
	type method struct{ fragment, material string }

	// docWithMethods serves a document whose verificationMethod set is exactly
	// the given methods. With declare unset the document carries no
	// relationships, like did:plc, and the validator tries the methods in
	// map order. With declare set, capabilityInvocation lists the methods in
	// the order given, and the validator tries them in that order.
	docWithMethods := func(declare bool, methods ...method) did.Resolver {
		return did.ResolverFunc(func(_ context.Context, d did.DID) (did.Document, error) {
			doc := did.NewDocument(d)
			for _, m := range methods {
				vm := did.VerificationMethod{
					ID:         doc.Fragment(m.fragment),
					Controller: d,
					Type:       did.MultikeyVerificationMethodType,
					Material:   did.GenericMap{did.MultikeyPublicKeyMultibaseProp: m.material},
				}
				var err error
				if declare {
					err = doc.CapabilityInvocation.Add(vm)
				} else {
					err = doc.VerificationMethods.Add(vm)
				}
				if err != nil {
					return did.Document{}, err
				}
			}
			return doc, nil
		})
	}

	t.Run("undecodable verification method does not veto a valid one", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		inv, err := invocation.Invoke(subject, subject.DID(), crankWidget, datamodel.Map{})
		require.NoError(t, err)

		// The declared relationship fixes the order: the undecodable method is
		// tried first, so the valid signer is only reached if the failure is
		// recorded rather than returned.
		ordered := docWithMethods(true,
			method{"wrap", undecodableKeyMultibase(t)},
			method{"signer", subject.DID().Identifier()},
		)
		err = validator.ValidateInvocation(t.Context(), inv,
			validator.WithDIDResolver(ordered))
		require.NoError(t, err)

		// The same document without relationships, as did:plc publishes it.
		// The order is unspecified here; the outcome must not depend on it.
		undeclared := docWithMethods(false,
			method{"wrap", undecodableKeyMultibase(t)},
			method{"signer", subject.DID().Identifier()},
		)
		err = validator.ValidateInvocation(t.Context(), inv,
			validator.WithDIDResolver(undeclared))
		require.NoError(t, err)
	})

	t.Run("context errors from a verifier factory are propagated", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		inv, err := invocation.Invoke(subject, subject.DID(), crankWidget, datamodel.Map{})
		require.NoError(t, err)

		resolver := docWithMethods(true,
			method{"remote", undecodableKeyMultibase(t)},
			method{"signer", subject.DID().Identifier()},
		)
		cancelled := func(context.Context, did.VerificationMaterial) (ucan.Verifier, error) {
			return nil, fmt.Errorf("fetching key: %w", context.Canceled)
		}

		err = validator.ValidateInvocation(t.Context(), inv,
			validator.WithDIDResolver(resolver),
			validator.WithVerifierFactories(map[string]validator.VerifierFactory{
				did.MultikeyVerificationMethodType: cancelled,
			}))
		require.ErrorIs(t, err, context.Canceled)
		require.NotContains(t, err.Error(), "does not have a valid signature")
	})

	t.Run("undecodable verification method is reported when nothing verifies", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		resolver := docWithMethods(false, method{"wrap", undecodableKeyMultibase(t)})

		inv, err := invocation.Invoke(subject, subject.DID(), crankWidget, datamodel.Map{})
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv,
			validator.WithDIDResolver(resolver))
		require.ErrorContains(t, err, "does not have a valid signature")
		require.ErrorContains(t, err, "#wrap: unusable verification material")
		require.ErrorContains(t, err, "no decoder registered for key type code")
	})

	t.Run("relationship referencing a missing method does not fall back", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)

		// The document lists the signer under verificationMethod, but
		// capabilityInvocation references an ID with no matching method. The
		// relationship is present (not silent), so verification must restrict
		// to it — not widen to all of verificationMethod.
		resolver := did.ResolverFunc(func(_ context.Context, d did.DID) (did.Document, error) {
			docJSON := fmt.Sprintf(`{
				"id": %q,
				"verificationMethod": [{
					"id": "%s#key",
					"type": %q,
					"controller": %q,
					"publicKeyMultibase": %q
				}],
				"capabilityInvocation": ["%s#missing"]
			}`, d, d, did.MultikeyVerificationMethodType, d, d.Identifier(), d)
			var doc did.Document
			if err := json.Unmarshal([]byte(docJSON), &doc); err != nil {
				return did.Document{}, err
			}
			return doc, nil
		})

		inv, err := invocation.Invoke(subject, subject.DID(), crankWidget, datamodel.Map{})
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv,
			validator.WithDIDResolver(resolver))
		require.Error(t, err)
	})

	t.Run("explicitly empty capabilityInvocation endorses no methods", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)

		// The document lists the signer under verificationMethod, but declares
		// an empty capabilityInvocation. Declaring nothing is not the same as
		// declaring nothing explicitly: the empty array endorses no methods, so
		// verification must not fall back to the full verificationMethod set.
		resolver := did.ResolverFunc(func(_ context.Context, d did.DID) (did.Document, error) {
			docJSON := fmt.Sprintf(`{
				"id": %q,
				"verificationMethod": [{
					"id": "%s#key",
					"type": %q,
					"controller": %q,
					"publicKeyMultibase": %q
				}],
				"capabilityInvocation": []
			}`, d, d, did.MultikeyVerificationMethodType, d, d.Identifier())
			var doc did.Document
			if err := json.Unmarshal([]byte(docJSON), &doc); err != nil {
				return did.Document{}, err
			}
			return doc, nil
		})

		inv, err := invocation.Invoke(subject, subject.DID(), crankWidget, datamodel.Map{})
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv,
			validator.WithDIDResolver(resolver))
		require.ErrorContains(t, err, "does not have a valid signature")
	})

	t.Run("explicitly empty capabilityDelegation endorses no methods", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)
		bob := testutil.RandomIssuer(t)

		// capabilityInvocation is absent (so the invocation itself verifies
		// against every method), while capabilityDelegation is explicitly empty
		// — isolating the proof's signature check, which must fail.
		resolver := did.ResolverFunc(func(_ context.Context, d did.DID) (did.Document, error) {
			docJSON := fmt.Sprintf(`{
				"id": %q,
				"verificationMethod": [{
					"id": "%s#key",
					"type": %q,
					"controller": %q,
					"publicKeyMultibase": %q
				}],
				"capabilityDelegation": []
			}`, d, d, did.MultikeyVerificationMethodType, d, d.Identifier())
			var doc did.Document
			if err := json.Unmarshal([]byte(docJSON), &doc); err != nil {
				return did.Document{}, err
			}
			return doc, nil
		})

		del, err := delegation.Delegate(subject, bob.DID(), subject.DID(), crankWidget)
		require.NoError(t, err)
		inv, err := invocation.Invoke(bob, subject.DID(), crankWidget, datamodel.Map{},
			invocation.WithProofs(del.Link()))
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv,
			validator.WithDIDResolver(resolver),
			validator.WithProofResolver(validator.ProofsFromContainer(
				container.New(container.WithDelegations(del)))),
		)
		require.ErrorContains(t, err, "does not have a valid signature")
	})

	t.Run("nil relationships fall back without panicking", func(t *testing.T) {
		subject := testutil.RandomIssuer(t)

		// A document built without NewDocument: relationship fields are nil.
		resolver := did.ResolverFunc(func(_ context.Context, d did.DID) (did.Document, error) {
			vms := did.VerificationMethods{}
			doc := did.Document{ID: d, VerificationMethods: &vms}
			vm := did.VerificationMethod{
				ID:         doc.Fragment("key"),
				Controller: d,
				Type:       did.MultikeyVerificationMethodType,
				Material:   did.GenericMap{did.MultikeyPublicKeyMultibaseProp: d.Identifier()},
			}
			if err := vms.Add(vm); err != nil {
				return did.Document{}, err
			}
			return doc, nil
		})

		inv, err := invocation.Invoke(subject, subject.DID(), crankWidget, datamodel.Map{})
		require.NoError(t, err)

		err = validator.ValidateInvocation(t.Context(), inv,
			validator.WithDIDResolver(resolver))
		require.NoError(t, err)
	})
}

// expiredKeyResolver returns a DID resolver that serves, for whichever DID it
// is asked to resolve, a document whose single Multikey VM is that DID's own
// key, marked expired or revoked as specified.
func expiredKeyResolver(t *testing.T, expires, revoked *did.DateTimeStamp) did.Resolver {
	t.Helper()
	return did.ResolverFunc(func(_ context.Context, d did.DID) (did.Document, error) {
		doc := did.NewDocument(d)
		vm := did.VerificationMethod{
			ID:         doc.Fragment(d.Identifier()),
			Controller: d,
			Expires:    expires,
			Revoked:    revoked,
			Type:       did.MultikeyVerificationMethodType,
			Material:   did.GenericMap{did.MultikeyPublicKeyMultibaseProp: d.Identifier()},
		}
		if err := doc.VerificationMethods.Add(vm); err != nil {
			return did.Document{}, err
		}
		for _, rel := range []*did.VerificationRelationship{
			doc.Authentication, doc.AssertionMethod,
			doc.CapabilityDelegation, doc.CapabilityInvocation,
		} {
			if err := rel.Add(vm); err != nil {
				return did.Document{}, err
			}
		}
		return doc, nil
	})
}

type StubVerifier struct {
	did did.DID
}

func (s StubVerifier) DID() did.DID {
	return s.did
}

func (s StubVerifier) Verify(msg []byte, sig []byte) bool {
	return false
}

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

			opts := []validator.Option{
				validator.WithValidationTime(ucan.UnixTimestamp(vector.Time)),
				validator.WithProofResolver(newMapProofResolver(proofs)),
			}

			err = validator.ValidateInvocation(t.Context(), inv, opts...)
			require.NoError(t, err, "validation should have passed for invocation with %s", vector.Description)
		})
	}

	for _, vector := range fixtures.Invalid {
		t.Run("invalid "+vector.Name, func(t *testing.T) {

			inv, err := invocation.Decode(vector.Invocation)
			require.NoError(t, err)
			t.Log("invocation", inv.Link())

			proofs := decodeProofs(t, vector.Proofs)

			opts := []validator.Option{
				validator.WithValidationTime(ucan.UnixTimestamp(vector.Time)),
				validator.WithProofResolver(newMapProofResolver(proofs)),
			}

			err = validator.ValidateInvocation(t.Context(), inv, opts...)
			require.Error(t, err, "validation should not have passed for invocation because %s", vector.Description)
			t.Log(err)

			var namedErr NamedError
			require.True(t, errors.As(err, &namedErr))
			require.Equal(t, vector.Error.Name, namedErr.Name())
		})
	}
}

func newMapProofResolver(proofs map[cid.Cid]ucan.Delegation) validator.ProofResolverFunc {
	return func(_ context.Context, link cid.Cid) (ucan.Delegation, error) {
		dlg, ok := proofs[link]
		if !ok {
			return nil, verrs.NewUnavailableProofError(link, errors.New("not provided"))
		}
		return dlg, nil
	}
}

func decodeProofs(t *testing.T, vectorProofs [][]byte) map[cid.Cid]ucan.Delegation {
	proofs := map[cid.Cid]ucan.Delegation{}
	for _, p := range vectorProofs {
		dlg, err := delegation.Decode(p)
		require.NoError(t, err)
		proofs[dlg.Link()] = dlg
		t.Log("proof", dlg.Link())
	}
	return proofs
}
