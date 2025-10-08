package validator

import (
	"fmt"
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

func NewExpiredError(t ucan.UCAN) vdm.ErrorModel {
	return vdm.ErrorModel{
		Name:    ExpiredErrorName,
		Message: fmt.Sprintf("proof %s has expired on %s", t.Link(), time.Unix(int64(*t.Expiration()), 0).Format(time.RFC3339)),
	}
}

const MalformedArgumentsError = "MalformedArguments"

func NewMalformedArgumentsError(cmd ucan.Command, cause error) vdm.ErrorModel {
	return vdm.ErrorModel{
		Name:    MalformedArgumentsError,
		Message: fmt.Sprintf("malformed arguments for command %s: %s", cmd, cause.Error()),
	}
}
