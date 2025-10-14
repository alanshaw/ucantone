package validator

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/principal"
	"github.com/alanshaw/ucantone/principal/verifier"
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

type validationConfig struct {
	canIssue              CanIssueFunc
	parsePrincipal        PrincipalParserFunc
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

func WithPrincipalParser(parsePrincipal PrincipalParserFunc) Option {
	return func(vc *validationConfig) {
		vc.parsePrincipal = parsePrincipal
	}
}

func WithProofResolver(resolveProof ProofResolverFunc) Option {
	return func(vc *validationConfig) {
		vc.resolveProof = resolveProof
	}
}

func WithDIDResolver(resolveDIDKey DIDResolverFunc) Option {
	return func(vc *validationConfig) {
		vc.resolveDIDKey = resolveDIDKey
	}
}

func WithAuthorizationValidator(validateAuthorization ValidateAuthorizationFunc) Option {
	return func(vc *validationConfig) {
		vc.validateAuthorization = validateAuthorization
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

	err = Validate(ctx, invocation, proofs)
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
func Validate(ctx context.Context, invocation ucan.Invocation, proofs map[cid.Cid]ucan.Delegation) error {
	err := ValidateNotExpired(invocation)
	if err != nil {
		return err
	}

	for _, p := range proofs {
		err := ValidateNotExpired(p)
		if err != nil {
			return err
		}
		err = ValidateNotTooEarly(p)
		if err != nil {
			return err
		}
	}

	return VerifyAuthorization()
}

func ValidateNotExpired(token ucan.Token) error {
	if ucan.IsExpired(token) {
		return NewExpiredError(token)
	}
	return nil
}

func ValidateNotTooEarly(dlg ucan.Delegation) error {
	if ucan.IsTooEarly(dlg) {
		return NewTooEarlyError(dlg)
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
			return NewUnverifiableSignatureError(inv, err)
		}
		return VerifySignature(inv, verifier)
	}

	if inv.Issuer().DID() == authority.DID() {
		return VerifySignature(inv, authority)
	}

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

		err = VerifySignature(inv, wvfr)
		if err != nil {
			verifyErr = err
			continue
		}
		break
	}

	if verifyErr != nil {
		return NewUnverifiableSignatureError(inv, verifyErr)
	}

	return nil
}

// VerifySignature verifies the token was signed by the passed verifier.
func VerifySignature(token ucan.Token, verifier ucan.Verifier) error {
	ok := ucan.VerifySignature(token, verifier)
	if !ok {
		return NewInvalidSignatureError(token, verifier)
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
func NopValidateAuthorization(ctx context.Context, auth Authorization[Arguments]) error {
	return nil
}
