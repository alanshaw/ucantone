package validator

import (
	"fmt"
	"strings"
	"time"

	"github.com/alanshaw/ucantone/did"
	"github.com/alanshaw/ucantone/ucan"
	vdm "github.com/alanshaw/ucantone/validator/datamodel"
)

const UnavailableProofErrorName = "UnavailableProof"

func NewUnavailableProofError(p ucan.Link, cause error) vdm.ErrorModel {
	return vdm.ErrorModel{
		Name:    UnavailableProofErrorName,
		Message: fmt.Sprintf("linked proof %s could not be resolved: %s", p.String(), cause.Error()),
	}
}

const DIDKeyResolutionErrorName = "DIDKeyResolutionError"

func NewDIDKeyResolutionError(d did.DID, cause error) vdm.ErrorModel {
	return vdm.ErrorModel{
		Name:    DIDKeyResolutionErrorName,
		Message: fmt.Sprintf("unable to resolve %s key: %s", d.String(), cause.Error()),
	}
}

const ExpiredErrorName = "Expired"

func NewExpiredError(t ucan.Token) vdm.ErrorModel {
	var name string
	if _, ok := t.(ucan.Invocation); ok {
		name = "invocation"
	} else {
		name = "proof"
	}
	return vdm.ErrorModel{
		Name:    ExpiredErrorName,
		Message: fmt.Sprintf("%s %s has expired on %s", name, t.Link(), time.Unix(int64(*t.Expiration()), 0).Format(time.RFC3339)),
	}
}

const TooEarlyErrorName = "TooEarly"

func NewTooEarlyError(t ucan.Delegation) vdm.ErrorModel {
	return vdm.ErrorModel{
		Name:    ExpiredErrorName,
		Message: fmt.Sprintf("proof %s is not valid before %s", t.Link(), time.Unix(int64(*t.NotBefore()), 0).Format(time.RFC3339)),
	}
}

const MalformedArgumentsErrorName = "MalformedArguments"

func NewMalformedArgumentsError(cmd ucan.Command, cause error) vdm.ErrorModel {
	return vdm.ErrorModel{
		Name:    MalformedArgumentsErrorName,
		Message: fmt.Sprintf("malformed arguments for command %s: %s", cmd, cause.Error()),
	}
}

const InvalidSignatureErrorName = "InvalidSignature"

func NewInvalidSignatureError(token ucan.Token, verifier ucan.Verifier) vdm.ErrorModel {
	issuer := token.Issuer().DID()
	key := verifier.DID()
	var message string
	if strings.HasPrefix(issuer.String(), "did:key") {
		message = fmt.Sprintf(`proof %s does not have a valid signature from %s`, token.Link(), key)
	} else {
		message = strings.Join([]string{
			fmt.Sprintf("proof %s issued by %s does not have a valid signature from %s", token.Link(), issuer, key),
			"  ℹ️ Issuer probably signed with a different key, which got rotated, invalidating delegations that were issued with prior keys",
		}, "\n")
	}
	return vdm.ErrorModel{
		Name:    InvalidSignatureErrorName,
		Message: message,
	}
}

const UnverifiableSignatureErrorName = "UnverifiableSignature"

func NewUnverifiableSignatureError(token ucan.Token, cause error) vdm.ErrorModel {
	issuer := token.Issuer().DID()
	return vdm.ErrorModel{
		Name:    UnverifiableSignatureErrorName,
		Message: fmt.Sprintf("proof %s issued by %s cannot be verified: %s", token.Link(), issuer, cause.Error()),
	}
}
