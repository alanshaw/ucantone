package validator

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/principal"
	edverifier "github.com/alanshaw/ucantone/principal/ed25519/verifier"
	"github.com/alanshaw/ucantone/principal/verifier"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/delegation"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
	"github.com/alanshaw/ucantone/ucan/invocation"
	"github.com/alanshaw/ucantone/validator/capability"
	verrs "github.com/alanshaw/ucantone/validator/errors"
	"github.com/ipfs/go-cid"
)

// Capability is a capability definition that can be used to validate an
// invocation and against it's proof policies.
type Capability interface {
	// Command is the command the capability matches against.
	Command() ucan.Command
	// Policy is the base policy for the capability.
	Policy() ucan.Policy
	// Match an invocation against the capability, resulting in a match, which is
	// the task from the invocation, verified to be matching with delegation
	// policies.
	Match(invocation ucan.Invocation, proofs map[cid.Cid]ucan.Delegation) (*capability.Match, error)
}

// Authorization is the details of an invocation that has been validated by the
// validator.
type Authorization struct {
	// Invocation is the invocation that was validated by the validator.
	Invocation ucan.Invocation
	// Proofs are the path of authority from the subject to the invoker. They are
	// delegations starting from the root Delegation (issued by the subject), in
	// strict sequence where the audience of the previous delegation matches the
	// issuer of the next Delegation.
	Proofs map[cid.Cid]ucan.Delegation
	Task   ucan.Task
}

// ProofResolverFunc finds a delegation corresponding to an external proof link.
type ProofResolverFunc func(ctx context.Context, link ucan.Link) (ucan.Delegation, error)

// CanIssueFunc determines whether given capability can be issued by a given
// principal or whether it needs to be delegated to the issuer.
type CanIssueFunc func(capability ucan.Capability, issuer ucan.Principal) bool

// ValidateAuthorizationFunc allows an authorization to be validated further. It
// is typically used to check that the delegations from the authorization have
// not been revoked. It returns `nil` on success.
type ValidateAuthorizationFunc func(ctx context.Context, auth Authorization) error

// DIDResolverFunc is used to resolve a key of the principal that is
// identified by DID different from did:key method. It can be passed into a
// UCAN validator in order to augment it with additional DID methods support.
type DIDResolverFunc func(ctx context.Context, nonDIDKey did.DID) ([]did.DID, error)

// PrincipalParserFunc provides verifier instances that can validate UCANs
// issued by a given principal.
type PrincipalParserFunc func(str string) (principal.Verifier, error)

// ValidationContext is the contextual information required by the validator in
// order to validate the delegation chain of an invocation.
type ValidationContext struct {
	// Authority is the identity of the local authority, used to verify signatures
	// of delegations signed by it.
	//
	// A capability provider service will use one corresponding to own DID or it's
	// supervisor's DID if it acts under it's authority.
	//
	// It also allows a service identified by non did:key e.g. did:web or did:dns
	// to pass a resolved key so it does not need to be resolved at runtime.
	Authority ucan.Verifier
	// CanIssue informs validator whether given capability can be issued by a
	// given principal or whether it needs to be delegated to the issuer. By
	// default, the validator will permit self signed invocations/delegations.
	CanIssue CanIssueFunc
	// ParsePrincipal provides verifier instances that can validate UCANs issued
	// by a given principal.
	ParsePrincipal PrincipalParserFunc
	// ResolveProof finds a delegation corresponding to a proof link.
	ResolveProof ProofResolverFunc
	// ResolveDIDKey is a function that resolves the key of a principal that is
	// identified by a DID method different from did:key.
	ResolveDIDKey DIDResolverFunc
	// ValidateAuthorization is called after an invocation has been validated to
	// allow an authorization to be validated further. It is typically used to
	// check that the delegations from the authorization have
	// not been revoked. It returns `nil` on success.
	ValidateAuthorization ValidateAuthorizationFunc
}

// Access validates the invocation issuer is authorized to invoke the delegated
// capability.
//
// The authority is the identity of the local authority, used to verify
// signatures of delegations signed by it.
//
// A capability provider service will use one corresponding to own DID or it's
// supervisor's DID if it acts under it's authority.
//
// It also allows a service identified by non did:key e.g. did:web or did:dns
// to pass a resolved key so it does not need to be resolved at runtime.
func Access(
	ctx context.Context,
	authority ucan.Verifier,
	capability Capability,
	invocation ucan.Invocation,
	options ...Option,
) (Authorization, error) {
	cfg := validationConfig{
		canIssue:              IsSelfIssued,
		parsePrincipal:        ParsePrincipal,
		resolveProof:          ProofUnavailable,
		resolveDIDKey:         FailDIDKeyResolution,
		validateAuthorization: NopValidateAuthorization,
	}
	for _, opt := range options {
		opt(&cfg)
	}

	proofs, err := ResolveProofs(ctx, cfg.resolveProof, invocation.Proofs())
	if err != nil {
		return Authorization{}, err
	}

	err = Validate(ctx, authority, cfg.canIssue, cfg.parsePrincipal, cfg.resolveDIDKey, invocation, proofs)
	if err != nil {
		return Authorization{}, err
	}

	match, err := capability.Match(invocation, proofs)
	if err != nil {
		return Authorization{}, err
	}

	return Authorization{
		Invocation: invocation,
		Task:       match.Task,
		Proofs:     match.Proofs,
	}, nil
}

func ResolveProofs(ctx context.Context, resolve ProofResolverFunc, links []ucan.Link) (map[cid.Cid]ucan.Delegation, error) {
	proofs := map[cid.Cid]ucan.Delegation{}
	for _, link := range links {
		prf, err := resolve(ctx, link)
		if err != nil {
			return nil, fmt.Errorf("resolving proof %s: %w", link.String(), err)
		}
		proofs[link] = prf
	}
	return proofs, nil
}

// Validate an invocation to check it is within the time bounds and that it is
// authorized by the issuer.
func Validate(
	ctx context.Context,
	authority ucan.Verifier,
	canIssue CanIssueFunc,
	parsePrincipal PrincipalParserFunc,
	resolveDIDKey DIDResolverFunc,
	inv ucan.Invocation,
	prfs map[cid.Cid]ucan.Delegation,
) error {
	err := ValidateNotExpired(inv)
	if err != nil {
		return err
	}

	for _, p := range prfs {
		err := ValidateNotExpired(p)
		if err != nil {
			return err
		}
		err = ValidateNotTooEarly(p)
		if err != nil {
			return err
		}
	}

	return VerifyAuthorization(ctx, authority, canIssue, parsePrincipal, resolveDIDKey, inv, prfs)
}

func ValidateNotExpired(token ucan.Token) error {
	if ucan.IsExpired(token) {
		return verrs.NewExpiredError(token)
	}
	return nil
}

func ValidateNotTooEarly(dlg ucan.Delegation) error {
	if ucan.IsTooEarly(dlg) {
		return verrs.NewTooEarlyError(dlg)
	}
	return nil
}

// VerifyAuthorization verifies that the invocation has been authorized by the
// issuer. If issued by the did:key principal it checks that the signature is
// valid. If issued by the root authority it checks that the signature is valid.
// If issued by the principal identified by other DID method attempts to resolve
// a valid `ucan/attest` attestation from the authority, if attestation is not
// found falls back to resolving did:key for the issuer and verifying its
// signature.
func VerifyAuthorization(
	ctx context.Context,
	authority ucan.Verifier,
	canIssue CanIssueFunc,
	parsePrincipal PrincipalParserFunc,
	resolveDIDKey DIDResolverFunc,
	inv ucan.Invocation,
	prfs map[cid.Cid]ucan.Delegation,
) error {
	issuer := inv.Issuer().DID()
	// If the issuer is a did:key we just verify a signature
	if strings.HasPrefix(issuer.String(), "did:key:") {
		verifier, err := parsePrincipal(issuer.String())
		if err != nil {
			return verrs.NewUnverifiableSignatureError(inv, err)
		}
		if err := VerifyInvocationSignature(inv, verifier); err != nil {
			return err
		}
	} else if inv.Issuer().DID() == authority.DID() {
		if err := VerifyInvocationSignature(inv, authority); err != nil {
			return err
		}
	} else {
		// TODO: verify attestations?

		// Otherwise we try to resolve did:key from the DID instead
		// and use that to verify the signature
		ids, err := resolveDIDKey(ctx, issuer)
		if err != nil {
			return err
		}

		var verifyErr error
		for _, id := range ids {
			vfr, err := parsePrincipal(id.String())
			if err != nil {
				verifyErr = err
				continue
			}
			wvfr, err := verifier.Wrap(vfr, issuer)
			if err != nil {
				verifyErr = err
				continue
			}
			err = VerifyInvocationSignature(inv, wvfr)
			if err != nil {
				verifyErr = err
				continue
			}
			break
		}
		if verifyErr != nil {
			return verrs.NewUnverifiableSignatureError(inv, verifyErr)
		}
	}

	if len(inv.Proofs()) > 0 {
		prf, ok := prfs[inv.Proofs()[0]]
		if !ok {
			return verrs.NewUnavailableProofError(inv.Proofs()[0], errors.New("missing from map"))
		}

		// check principal alignment
		if inv.Issuer().DID() != prf.Audience().DID() {
			return verrs.NewPrincipalAlignmentError(inv.Issuer(), prf)
		}

		for i, p := range inv.Proofs() {
			var next ucan.Delegation
			if i+1 < len(inv.Proofs()) {
				np, ok := prfs[inv.Proofs()[i+1]]
				if !ok {
					return verrs.NewUnavailableProofError(inv.Proofs()[i+1], errors.New("missing from map"))
				}
				next = np
			}

			prf, ok := prfs[p]
			if !ok {
				return verrs.NewUnavailableProofError(p, errors.New("missing from map"))
			}
			issuer := prf.Issuer().DID()

			// If the issuer is a did:key we just verify a signature
			if strings.HasPrefix(issuer.String(), "did:key:") {
				verifier, err := parsePrincipal(issuer.String())
				if err != nil {
					return verrs.NewUnverifiableSignatureError(prf, err)
				}
				if err := VerifyDelegationSignature(prf, verifier); err != nil {
					return err
				}
			} else if prf.Issuer().DID() == authority.DID() {
				if err := VerifyDelegationSignature(prf, authority); err != nil {
					return err
				}
			} else {
				// Otherwise we try to resolve did:key from the DID instead
				// and use that to verify the signature
				ids, err := resolveDIDKey(ctx, issuer)
				if err != nil {
					return err
				}

				var verifyErr error
				for _, id := range ids {
					vfr, err := parsePrincipal(id.String())
					if err != nil {
						verifyErr = err
						continue
					}
					wvfr, err := verifier.Wrap(vfr, issuer)
					if err != nil {
						verifyErr = err
						continue
					}
					err = VerifyDelegationSignature(prf, wvfr)
					if err != nil {
						verifyErr = err
						continue
					}
					break
				}
				if verifyErr != nil {
					return verrs.NewUnverifiableSignatureError(prf, verifyErr)
				}
			}

			// this is the root delegation
			if next == nil {
				// powerline is not allowed as root delegation.
				// a priori there is no such thing as a null subject.
				if prf.Subject() == nil {
					return verrs.NewInvalidClaimError("root delegation subject is null")
				}
				if prf.Subject().DID() != inv.Subject().DID() {
					return verrs.NewSubjectAlignmentError(inv.Subject(), prf)
				}
				// check root issuer/subject alignment
				if !canIssue(ucan.Capability(prf), prf.Issuer()) {
					return verrs.NewInvalidClaimError(fmt.Sprintf("%s cannot issue delegations for %s", prf.Issuer().DID(), prf.Subject().DID()))
				}
			} else {
				// otherwise check subject and principal alignment
				if prf.Subject() != nil && prf.Subject().DID() != inv.Subject().DID() {
					return verrs.NewSubjectAlignmentError(inv.Subject(), prf)
				}
				if prf.Issuer().DID() != next.Audience().DID() {
					return verrs.NewPrincipalAlignmentError(prf.Issuer(), next)
				}
			}
		}
	} else {
		// check invocation issuer/subject alignment
		cap := delegation.NewCapability(inv.Subject(), inv.Command(), policy.Policy{})
		if !canIssue(cap, inv.Issuer()) {
			return verrs.NewInvalidClaimError(fmt.Sprintf("%s cannot issue invocations for %s", inv.Issuer().DID(), inv.Subject().DID()))
		}
	}

	return nil
}

// VerifyInvocationSignature verifies the invocation was signed by the passed verifier.
func VerifyInvocationSignature(inv ucan.Invocation, verifier ucan.Verifier) error {
	ok, err := invocation.VerifySignature(inv, verifier)
	if err != nil {
		return err
	}
	if !ok {
		return verrs.NewInvalidSignatureError(inv, verifier)
	}
	return nil
}

// VerifyDelegationSignature verifies the delegation was signed by the passed verifier.
func VerifyDelegationSignature(dlg ucan.Delegation, verifier ucan.Verifier) error {
	ok, err := delegation.VerifySignature(dlg, verifier)
	if err != nil {
		return err
	}
	if !ok {
		return verrs.NewInvalidSignatureError(dlg, verifier)
	}
	return nil
}

// IsSelfIssued is a [CanIssueFunc] that allows delegations to be self signed.
func IsSelfIssued(capability ucan.Capability, issuer ucan.Principal) bool {
	return capability.Subject().DID() == issuer.DID()
}

// ParsePrincipal is a [PrincipalParser] that supports parsing ed25519 DIDs.
func ParsePrincipal(str string) (principal.Verifier, error) {
	return edverifier.Parse(str)
}

// ProofUnavailable is a [ProofResolverFunc] that always fails.
func ProofUnavailable(ctx context.Context, p ucan.Link) (ucan.Delegation, error) {
	return nil, verrs.NewUnavailableProofError(p, errors.New("no proof resolver configured"))
}

// FailDIDKeyResolution is a [DIDResolverFunc] that always fails.
func FailDIDKeyResolution(ctx context.Context, d did.DID) ([]did.DID, error) {
	return []did.DID{}, verrs.NewDIDKeyResolutionError(d, errors.New("no DID resolver configured"))
}

// NopValidateAuthorization is a [ValidateAuthorizationFunc] that does no
// validation and returns nil.
func NopValidateAuthorization(ctx context.Context, auth Authorization) error {
	return nil
}
