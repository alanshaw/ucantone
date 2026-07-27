package did

import (
	"encoding/json"
	"errors"
)

type VerificationRelationship struct {
	allMethods          *VerificationMethods
	relationshipMethods []URL
	// declared records whether this relationship was explicitly stated —
	// present in the document JSON (even as an empty array) or built up via
	// Add. An undeclared relationship endorses every method in the document
	// (the default Document.UnmarshalJSON establishes); a declared one
	// endorses exactly its references.
	declared bool
}

func (vr *VerificationRelationship) Add(method VerificationMethod) error {
	if vr.allMethods == nil {
		return errors.New("verification relationship is not associated with a document")
	}
	if err := vr.allMethods.Add(method); err != nil {
		return err
	}
	vr.relationshipMethods = append(vr.relationshipMethods, method.ID)
	vr.declared = true
	return nil
}

// All returns the verification methods this relationship endorses. An
// undeclared relationship (no such section in the document — e.g. every
// document the PLC directory serves) falls back to all of the document's
// methods; a declared relationship, even an empty one, yields exactly its
// references.
func (vr *VerificationRelationship) All() []VerificationMethod {
	if !vr.declared {
		if vr.allMethods == nil {
			return nil
		}
		return vr.allMethods.All()
	}
	vms := make([]VerificationMethod, 0, len(vr.relationshipMethods))
	for _, u := range vr.relationshipMethods {
		if vm, ok := (*vr.allMethods)[u.String()]; ok {
			vms = append(vms, vm)
		}
	}
	return vms
}

func (vr *VerificationRelationship) Get(i int) URL {
	return vr.relationshipMethods[i]
}

func (vr *VerificationRelationship) Len() int {
	return len(vr.relationshipMethods)
}

func (vr *VerificationRelationship) IsZero() bool {
	return len(vr.relationshipMethods) == 0
}

func (vr *VerificationRelationship) MarshalJSON() ([]byte, error) {
	return json.Marshal(vr.relationshipMethods)
}

func (vr *VerificationRelationship) UnmarshalJSON(data []byte) error {
	var raws []json.RawMessage
	err := json.Unmarshal(data, &raws)
	if err != nil {
		return err
	}
	vr.declared = true

	for _, raw := range raws {
		var u URL
		err := json.Unmarshal(raw, &u)
		if err == nil {
			vr.relationshipMethods = append(vr.relationshipMethods, u)
			continue
		}
		var typeErr *json.UnmarshalTypeError
		if !errors.As(err, &typeErr) {
			return err
		}
		var vm VerificationMethod
		if err := json.Unmarshal(raw, &vm); err != nil {
			return err
		}
		if err := vr.allMethods.Add(vm); err != nil {
			return err
		}
		vr.relationshipMethods = append(vr.relationshipMethods, vm.ID)
	}
	return nil
}
