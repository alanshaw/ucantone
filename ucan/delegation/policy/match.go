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
		selectedValue, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		statementValue := s.model.Value.Value
		if !reflect.DeepEqual(statementValue, selectedValue) {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" does not equal "%v"`, s.Selector(), selectedValue, statementValue))
		}
		return true, nil
	case OpNotEqual:
		selectedValue, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		statementValue := s.model.Value.Value
		if reflect.DeepEqual(statementValue, selectedValue) {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" equals "%v"`, s.Selector(), selectedValue, statementValue))
		}
		return true, nil
	case OpGreaterThan:
		selectedValue, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		if selectedValue == nil {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector(), selectedValue, OpGreaterThan))
		}
		if !isOrdered(selectedValue, s.model.Value.Value, gt) {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not greater than "%v"`, s.Selector(), selectedValue, s.model.Value.Value))
		}
		return true, nil
	case OpGreaterThanOrEqual:
		selectedValue, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		if selectedValue == nil {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector(), selectedValue, OpGreaterThanOrEqual))
		}
		if !isOrdered(selectedValue, s.model.Value.Value, gte) {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not greater than or equal to "%v"`, s.Selector(), selectedValue, s.model.Value.Value))
		}
		return true, nil
	case OpLessThan:
		selectedValue, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		if selectedValue == nil {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector(), selectedValue, OpLessThan))
		}
		if !isOrdered(selectedValue, s.model.Value.Value, lt) {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not less than "%v"`, s.Selector(), selectedValue, s.model.Value.Value))
		}
		return true, nil
	case OpLessThanOrEqual:
		selectedValue, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		if selectedValue == nil {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector(), selectedValue, OpLessThanOrEqual))
		}
		if !isOrdered(selectedValue, s.model.Value.Value, lte) {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not less than or equal to "%v"`, s.Selector(), selectedValue, s.model.Value.Value))
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
		return false, fmt.Errorf(`"%v" did not match any statements`, value)
	case OpLike:
		selectedValue, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		v, ok := selectedValue.(string)
		if !ok {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector(), selectedValue, OpLike))
		}
		if !s.glob.Match(v) {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not like "%v"`, s.Selector(), selectedValue, s.model.Pattern))
		}
		return true, nil
	case OpAll:
		selectedValue, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		selectedValueVal := reflect.ValueOf(selectedValue)
		if selectedValueVal.Kind() != reflect.Slice {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not a list`, s.Selector(), selectedValue))
		}
		for i := range selectedValueVal.Len() {
			ss := make([]Statement, 0, len(s.statements))
			for _, m := range s.statements {
				ss = append(ss, *m)
			}
			ok, err := Match(Policy{ss}, selectedValueVal.Index(i).Interface())
			if !ok {
				return false, err
			}
		}
		return true, nil
	case OpAny:
		selectedValue, err := selector.Select(s.selector, value)
		if err != nil {
			return false, err
		}
		selectedValueVal := reflect.ValueOf(selectedValue)
		if selectedValueVal.Kind() != reflect.Slice {
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not a list`, s.Selector(), selectedValue))
		}
		for i := range selectedValueVal.Len() {
			ss := make([]Statement, 0, len(s.statements))
			for _, m := range s.statements {
				ss = append(ss, *m)
			}
			ok, _ := Match(Policy{ss}, selectedValueVal.Index(i).Interface())
			if ok {
				return true, nil
			}
		}
		return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" did not match any statements`, s.Selector(), selectedValue))
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
