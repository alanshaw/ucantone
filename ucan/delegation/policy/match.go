package policy

import (
	"cmp"
	"fmt"
	"reflect"

	"github.com/alanshaw/ucantone/ucan/delegation/policy/selector"
)

// Match determines if the value matches the policy document. If the value fails
// to match, the returned error will contain details of the failure.
func Match(policy Policy, value any) (bool, error) {
	for _, stmt := range policy.Statements {
		ok, err := MatchStatement(stmt, value)
		if !ok {
			return ok, err
		}
	}
	return true, nil
}

func MatchStatement(statement Statement, value any) (bool, error) {
	switch statement.Operation() {
	case OpEqual:
		if s, ok := statement.(ComparisonStatement); ok {
			one, _, err := selector.Select(s.Selector, value)
			if err != nil {
				return false, err
			}
			if one == nil {
				// TODO: should be allowed
				return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector, one, OpEqual))
			}
			if !reflect.DeepEqual(s.Value, one) {
				return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" does not equal "%v"`, s.Selector, one, s.Value))
			}
			return true, nil
		}
	case OpGreaterThan:
		if s, ok := statement.(ComparisonStatement); ok {
			one, _, err := selector.Select(s.Selector, value)
			if err != nil {
				return false, err
			}
			if one == nil {
				return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector, one, OpGreaterThan))
			}
			if !isOrdered(one, s.Value, gt) {
				return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not greater than "%v"`, s.Selector, one, s.Value))
			}
			return true, nil
		}
	case OpGreaterThanOrEqual:
		if s, ok := statement.(ComparisonStatement); ok {
			one, _, err := selector.Select(s.Selector, value)
			if err != nil {
				return false, err
			}
			if one == nil {
				return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector, one, OpGreaterThanOrEqual))
			}
			if !isOrdered(one, s.Value, gte) {
				return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not greater than or equal to "%v"`, s.Selector, one, s.Value))
			}
			return true, nil
		}
	case OpLessThan:
		if s, ok := statement.(ComparisonStatement); ok {
			one, _, err := selector.Select(s.Selector, value)
			if err != nil {
				return false, err
			}
			if one == nil {
				return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector, one, OpLessThan))
			}
			if !isOrdered(one, s.Value, lt) {
				return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not less than "%v"`, s.Selector, one, s.Value))
			}
			return true, nil
		}
	case OpLessThanOrEqual:
		if s, ok := statement.(ComparisonStatement); ok {
			one, _, err := selector.Select(s.Selector, value)
			if err != nil {
				return false, err
			}
			if one == nil {
				return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector, one, OpLessThanOrEqual))
			}
			if !isOrdered(one, s.Value, lte) {
				return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not less than or equal to "%v"`, s.Selector, one, s.Value))
			}
			return true, nil
		}
	case OpNot:
		if s, ok := statement.(NegationStatement); ok {
			ok, _ := MatchStatement(s.Statement, value)
			return !ok, nil
		}
	case OpAnd:
		if s, ok := statement.(ConjunctionStatement); ok {
			for _, cs := range s.Statements {
				ok, err := MatchStatement(cs, value)
				if !ok {
					return false, err
				}
			}
			return true, nil
		}
	case OpOr:
		if s, ok := statement.(DisjunctionStatement); ok {
			if len(s.Statements) == 0 {
				return true, nil
			}
			for _, cs := range s.Statements {
				ok, _ := MatchStatement(cs, value)
				if ok {
					return true, nil
				}
			}
			return false, nil
		}
	case OpLike:
		if s, ok := statement.(WildcardStatement); ok {
			one, _, err := selector.Select(s.Selector, value)
			if err != nil {
				return false, err
			}
			v, ok := one.(string)
			if !ok {
				return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not applicable to operator "%s"`, s.Selector, one, OpLike))
			}
			if s.Glob.Match(v) {
				return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is not like "%v"`, s.Selector, one, s.Pattern))
			}
			return true, nil
		}
	case OpAll:
		if s, ok := statement.(QuantificationStatement); ok {
			_, many, err := selector.Select(s.Selector, value)
			if err != nil {
				return false, err
			}
			if many == nil {
				return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is empty or not a list`, s.Selector, many))
			}
			for _, n := range many {
				ok, err := Match(Policy{s.Statements}, n)
				if !ok {
					return false, err
				}
			}
			return true, nil
		}
	case OpAny:
		if s, ok := statement.(QuantificationStatement); ok {
			_, many, err := selector.Select(s.Selector, value)
			if err != nil {
				return false, err
			}
			for _, n := range many {
				ok, _ := Match(Policy{s.Statements}, n)
				if ok {
					return true, nil
				}
			}
			return false, NewMatchError(statement, fmt.Errorf(`matching "%s": "%v" is empty or not a list`, s.Selector, many))
		}
	}
	panic(fmt.Errorf("unknown statement operation: %s", statement.Operation()))
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
