package policy

import (
	"cmp"
	"fmt"
	"reflect"

	"github.com/alanshaw/ucantone/ucan/delegation/policy/selector"
)

// Match determines if the value matches the policy document.
func Match(policy Policy, value any) bool {
	for _, stmt := range policy.Statements {
		ok := MatchStatement(stmt, value)
		if !ok {
			return false
		}
	}
	return true
}

func MatchStatement(statement Statement, value any) bool {
	switch statement.Operation() {
	case OpEqual:
		if s, ok := statement.(ComparisonStatement); ok {
			one, _, err := selector.Select(s.Selector, value)
			if err != nil || one == nil {
				return false
			}
			return reflect.DeepEqual(s.Value, one)
		}
	case OpGreaterThan:
		if s, ok := statement.(ComparisonStatement); ok {
			one, _, err := selector.Select(s.Selector, value)
			if err != nil || one == nil {
				return false
			}
			return isOrdered(one, s.Value, gt)
		}
	case OpGreaterThanOrEqual:
		if s, ok := statement.(ComparisonStatement); ok {
			one, _, err := selector.Select(s.Selector, value)
			if err != nil || one == nil {
				return false
			}
			return isOrdered(one, s.Value, gte)
		}
	case OpLessThan:
		if s, ok := statement.(ComparisonStatement); ok {
			one, _, err := selector.Select(s.Selector, value)
			if err != nil || one == nil {
				return false
			}
			return isOrdered(one, s.Value, lt)
		}
	case OpLessThanOrEqual:
		if s, ok := statement.(ComparisonStatement); ok {
			one, _, err := selector.Select(s.Selector, value)
			if err != nil || one == nil {
				return false
			}
			return isOrdered(one, s.Value, lte)
		}
	case OpNot:
		if s, ok := statement.(NegationStatement); ok {
			return !MatchStatement(s.Statement, value)
		}
	case OpAnd:
		if s, ok := statement.(ConjunctionStatement); ok {
			for _, cs := range s.Statements {
				r := MatchStatement(cs, value)
				if !r {
					return false
				}
			}
			return true
		}
	case OpOr:
		if s, ok := statement.(DisjunctionStatement); ok {
			if len(s.Statements) == 0 {
				return true
			}
			for _, cs := range s.Statements {
				r := MatchStatement(cs, value)
				if r {
					return true
				}
			}
			return false
		}
	case OpLike:
		if s, ok := statement.(WildcardStatement); ok {
			one, _, err := selector.Select(s.Selector, value)
			if err != nil || one == nil {
				return false
			}
			if v, ok := one.(string); ok {
				return s.Glob.Match(v)
			}
			return false
		}
	case OpAll:
		if s, ok := statement.(QuantificationStatement); ok {
			_, many, err := selector.Select(s.Selector, value)
			if err != nil || many == nil {
				return false
			}
			for _, n := range many {
				ok := Match(Policy{s.Statements}, n)
				if !ok {
					return false
				}
			}
			return true
		}
	case OpAny:
		if s, ok := statement.(QuantificationStatement); ok {
			_, many, err := selector.Select(s.Selector, value)
			if err != nil || many == nil {
				return false
			}
			for _, n := range many {
				ok := Match(Policy{s.Statements}, n)
				if ok {
					return true
				}
			}
			return false
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
