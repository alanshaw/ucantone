package policy

import (
	"cmp"
	"errors"
	"fmt"
	"reflect"

	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/delegation/policy/selector"
)

// Match determines if the value matches the policy document. If the value fails
// to match, the returned error will contain details of the failure.
func Match(policy ucan.Policy, value any) (bool, error) {
	for _, stmt := range policy.Statements() {
		ok, err := MatchStatement(stmt, value)
		if !ok {
			return ok, err
		}
	}
	return true, nil
}

func MatchStatement(statement ucan.Statement, value any) (bool, error) {
	s, err := toStatement(statement)
	if err != nil {
		return false, err
	}
	switch statement.Operator() {
	case OpEqual:
		one, _, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		if one == nil {
			// TODO: should be allowed
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector(), one, OpEqual))
		}
		if !reflect.DeepEqual(s.model.Value.Value, one) {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" does not equal "%v"`, s.Selector(), one, s.model.Value.Value))
		}
		return true, nil
	case OpGreaterThan:
		one, _, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		if one == nil {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector(), one, OpGreaterThan))
		}
		if !isOrdered(one, s.model.Value.Value, gt) {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not greater than "%v"`, s.Selector(), one, s.model.Value.Value))
		}
		return true, nil
	case OpGreaterThanOrEqual:
		one, _, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		if one == nil {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector(), one, OpGreaterThanOrEqual))
		}
		if !isOrdered(one, s.model.Value.Value, gte) {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not greater than or equal to "%v"`, s.Selector(), one, s.model.Value.Value))
		}
		return true, nil
	case OpLessThan:
		one, _, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		if one == nil {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector(), one, OpLessThan))
		}
		if !isOrdered(one, s.model.Value.Value, lt) {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not less than "%v"`, s.Selector(), one, s.model.Value.Value))
		}
		return true, nil
	case OpLessThanOrEqual:
		one, _, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		if one == nil {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector(), one, OpLessThanOrEqual))
		}
		if !isOrdered(one, s.model.Value.Value, lte) {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not less than or equal to "%v"`, s.Selector(), one, s.model.Value.Value))
		}
		return true, nil
	case OpNot:
		ss, ok := statement.Argument().(ucan.Statement)
		if !ok {
			return false, fmt.Errorf(`"%s" operator argument is not a statement`, s.Operator())
		}
		ok, _ = MatchStatement(ss, value)
		if ok {
			return false, NewMatchError(statement, errors.New("not true is false"))
		}
		return true, nil
	case OpAnd:
		for _, s := range s.statements {
			ok, err := MatchStatement(s, value)
			if !ok {
				return false, err
			}
		}
		return true, nil
	case OpOr:
		if len(s.statements) == 0 {
			return true, nil
		}
		for _, s := range s.statements {
			ok, _ := MatchStatement(s, value)
			if ok {
				return true, nil
			}
		}
		return false, nil
	case OpLike:
		one, _, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		v, ok := one.(string)
		if !ok {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector(), one, OpLike))
		}
		if s.glob.Match(v) {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not like "%v"`, s.Selector(), one, s.model.Pattern))
		}
		return true, nil
	case OpAll:
		_, many, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		if many == nil {
			return false, NewMatchError(statement, fmt.Errorf(`"%v" is empty or not a list`, many))
		}
		for _, n := range many {
			ss := make([]Statement, 0, len(s.statements))
			for _, m := range s.statements {
				ss = append(ss, *m)
			}
			ok, err := Match(Policy{ss}, n)
			if !ok {
				return false, err
			}
		}
		return true, nil
	case OpAny:
		_, many, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		for _, n := range many {
			ss := make([]Statement, 0, len(s.statements))
			for _, m := range s.statements {
				ss = append(ss, *m)
			}
			ok, _ := Match(Policy{ss}, n)
			if ok {
				return true, nil
			}
		}
		return false, NewMatchError(statement, fmt.Errorf(`"%v" is empty or not a list`, many))
	}
	panic(fmt.Errorf("unknown statement operator: %s", statement.Operator()))
}

func isOrdered(a any, b any, satisfies func(order int) bool) bool {
	if aint, ok := a.(int); ok {
		a = int64(aint)
	}
	if bint, ok := b.(int); ok {
		b = int64(bint)
	}
	if aint64, ok := a.(int64); ok {
		if bint64, ok := b.(int64); ok {
			return satisfies(cmp.Compare(aint64, bint64))
		}
	}
	// TODO: support float
	return false
}

func gt(order int) bool  { return order == 1 }
func gte(order int) bool { return order == 0 || order == 1 }
func lt(order int) bool  { return order == -1 }
func lte(order int) bool { return order == 0 || order == -1 }
