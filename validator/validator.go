package validator

import (
	"context"
	"errors"
	"fmt"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ucan"
	"github.com/ipfs/go-cid"
)

// Authorization is the details of an invocation that has been validated by the
// validator.
type Authorization[A Arguments] struct {
	// Invocation is the invocation that was validated by the validator.
	Invocation ucan.Invocation
	// Proofs are the path of authority from the subject to the invoker. They are
	// delegations starting from the root Delegation (issued by the subject), in
	// strict sequence where the audience of the previous delegation matches the
	// issuer of the next Delegation.
	Proofs map[cid.Cid]ucan.Delegation
	Task   *Task[A]
}

// ProofResolverFunc finds a delegation corresponding to an external proof link.
type ProofResolverFunc func(ctx context.Context, link ucan.Link) (ucan.Delegation, error)

// CanIssueFunc determines whether given capability can be issued by a given
// principal or whether it needs to be delegated to the issuer.
type CanIssueFunc func(capability ucan.Capability, issuer ucan.Principal) bool

// ValidateAuthorizationFunc allows an authorization to be validated further. It
// is typically used to check that the delegations from the authorization have
// not been revoked. It returns `nil` on success.
type ValidateAuthorizationFunc func(ctx context.Context, auth Authorization[Arguments]) error

// DIDResolverFunc is used to resolve a key of the principal that is
// identified by DID different from did:key method. It can be passed into a
// UCAN validator in order to augment it with additional DID methods support.
type DIDResolverFunc func(ctx context.Context, nonDIDKey did.DID) ([]did.DID, error)

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

type validationConfig struct {
	canIssue              CanIssueFunc
	resolveProof          ProofResolverFunc
	resolveDIDKey         DIDResolverFunc
	validateAuthorization ValidateAuthorizationFunc
}

// Option is an option configuring the validator.
type Option func(*validationConfig)

// WithCanIssue informs validator whether given capability can be issued by a
// given principal or whether it needs to be delegated to the issuer.
func WithCanIssue(canIssue CanIssueFunc) Option {
	return func(vc *validationConfig) {
		vc.canIssue = canIssue
	}
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
func Access[A Arguments](
	ctx context.Context,
	authority ucan.Verifier,
	capability *Capability[A],
	invocation ucan.Invocation,
	options ...Option,
) (Authorization[A], error) {
	cfg := validationConfig{
		canIssue:              IsSelfIssued,
		resolveProof:          ProofUnavailable,
		resolveDIDKey:         FailDIDKeyResolution,
		validateAuthorization: NopValidateAuthorization,
	}
	for _, opt := range options {
		opt(&cfg)
	}

	proofs, err := ResolveProofs(ctx, cfg.resolveProof, invocation.Proofs())
	if err != nil {
		return Authorization[A]{}, err
	}

	match, err := capability.Match(invocation, proofs)
	if err != nil {
		return Authorization[A]{}, err
	}

	return Authorization[A]{
		Invocation: invocation,
		Task:       match.Value,
		Proofs:     proofs,
	}, nil
}

func ResolveProofs(ctx context.Context, resolve ProofResolverFunc, links []ucan.Link) (map[cid.Cid]ucan.Delegation, error) {
	var proofs map[cid.Cid]ucan.Delegation
	for _, link := range links {
		prf, err := resolve(ctx, link)
		if err != nil {
			return nil, fmt.Errorf("resolving proof %s: %w", link.String(), err)
		}
		proofs[link] = prf
	}
	return proofs, nil
}

// Validate an invocation to check it is within the time bound and that it is
// authorized by the issuer.
func ValidateTimeBounds(ctx context.Context, token ucan.UCAN) InvalidProof {
	if ucan.IsExpired(token) {
		return NewExpiredError(token)
	}
	if dlg, ok := token.(ucan.Delegation); ok {
		if ucan.IsTooEarly(dlg) {
			return NewNotValidBeforeError(token)
		}
	}
	return nil
}

// IsSelfIssued is a [CanIssueFunc] that allows delegations to be self signed.
func IsSelfIssued(capability ucan.Capability, issuer ucan.Principal) bool {
	return capability.Subject().DID() == issuer.DID()
}

// ProofUnavailable is a [ProofResolverFunc] that always fails.
func ProofUnavailable(ctx context.Context, p ucan.Link) (ucan.Delegation, error) {
	return nil, NewUnavailableProofError(p, errors.New("no proof resolver configured"))
}

// FailDIDKeyResolution is a [DIDResolverFunc] that always fails.
func FailDIDKeyResolution(ctx context.Context, d did.DID) ([]did.DID, error) {
	return []did.DID{}, NewDIDKeyResolutionError(d, errors.New("no DID resolver configured"))
}

// NopValidateAuthorization is a [ValidateAuthorizationFunc] that does no
// validation and returns nil.
func NopValidateAuthorization(ctx context.Context, auth Authorization) error {
	return nil
}
