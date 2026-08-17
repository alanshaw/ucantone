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

// IsZero reports whether this relationship was never declared, and so is the
// `omitzero` predicate for the [Document] relationship fields: an undeclared
// relationship is absent from the document JSON, while a declared one is
// emitted even when it endorses nothing. It is deliberately NOT "endorses no
// methods" — that would marshal a declared empty relationship as an absent
// one, widening it from endorsing nothing to endorsing everything.
func (vr *VerificationRelationship) IsZero() bool {
	return !vr.declared
}

func (vr *VerificationRelationship) MarshalJSON() ([]byte, error) {
	if !vr.declared {
		// Unreachable via Document marshaling (omitzero omits undeclared
		// relationships), but a standalone marshal must emit null — the value
		// UnmarshalJSON maps back to undeclared — not "[]", which would
		// round-trip as declared-empty and endorse nothing.
		return []byte("null"), nil
	}
	if vr.relationshipMethods == nil {
		// A declared relationship with no references is an empty array, not
		// null: null unmarshals back as undeclared.
		return []byte("[]"), nil
	}
	return json.Marshal(vr.relationshipMethods)
}

func (vr *VerificationRelationship) UnmarshalJSON(data []byte) error {
	// By convention an unmarshaler treats null as a no-op, and here that is
	// also the right semantics: a null relationship declares nothing, so it
	// falls back to all of the document's methods rather than endorsing none.
	if string(data) == "null" {
		return nil
	}

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
