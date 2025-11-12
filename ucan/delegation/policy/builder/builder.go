package builder

import (
	"fmt"

	"github.com/alanshaw/ucantone/ucan"
	"github.com/alanshaw/ucantone/ucan/delegation/policy"
)

type Statement struct {
	Operator string
	Selector string
	Argument any
}

func Build(statements ...Statement) (ucan.Policy, error) {
	bstmts, err := BuildStatements(statements...)
	if err != nil {
		return policy.Policy{}, fmt.Errorf("building policy statements: %w", err)
	}
	return policy.New(bstmts...)
}

func BuildStatement(s Statement) (ucan.Statement, error) {
	switch s.Operator {
	case policy.OpEqual:
		return policy.Equal(s.Selector, s.Argument)
	case policy.OpNotEqual:
		return policy.NotEqual(s.Selector, s.Argument)
	case policy.OpGreaterThan:
		return policy.GreaterThan(s.Selector, s.Argument)
	case policy.OpGreaterThanOrEqual:
		return policy.GreaterThanOrEqual(s.Selector, s.Argument)
	case policy.OpLessThan:
		return policy.LessThan(s.Selector, s.Argument)
	case policy.OpLessThanOrEqual:
		return policy.LessThanOrEqual(s.Selector, s.Argument)
	case policy.OpAnd:
		arg, ok := s.Argument.([]Statement)
		if !ok {
			return nil, fmt.Errorf(`building "%s" statement: argument is not a statement list`, s.Operator)
		}
		sss, err := BuildStatements(arg...)
		if err != nil {
			return nil, fmt.Errorf(`building "%s" argument statements: %w`, s.Operator, err)
		}
		return policy.And(sss...)
	case policy.OpOr:
		arg, ok := s.Argument.([]Statement)
		if !ok {
			return nil, fmt.Errorf(`building "%s" statement: argument is not a statement list`, s.Operator)
		}
		sss, err := BuildStatements(arg...)
		if err != nil {
			return nil, fmt.Errorf(`building "%s" argument statements: %w`, s.Operator, err)
		}
		return policy.Or(sss...)
	case policy.OpAll:
		arg, ok := s.Argument.([]Statement)
		if !ok {
			return nil, fmt.Errorf(`building "%s" statement: argument is not a statement list`, s.Operator)
		}
		sss, err := BuildStatements(arg...)
		if err != nil {
			return nil, fmt.Errorf(`building "%s" argument statements: %w`, s.Operator, err)
		}
		return policy.All(s.Selector, sss...)
	case policy.OpAny:
		arg, ok := s.Argument.([]Statement)
		if !ok {
			return nil, fmt.Errorf(`building "%s" statement: argument is not a statement list`, s.Operator)
		}
		sss, err := BuildStatements(arg...)
		if err != nil {
			return nil, fmt.Errorf(`building "%s" argument statements: %w`, s.Operator, err)
		}
		return policy.Any(s.Selector, sss...)
	case policy.OpNot:
		arg, ok := s.Argument.(Statement)
		if !ok {
			return nil, fmt.Errorf(`building "%s" statement: argument is not a statement`, s.Operator)
		}
		ss, err := BuildStatement(arg)
		if err != nil {
			return nil, fmt.Errorf(`building "%s" argument statement: %w`, s.Operator, err)
		}
		return policy.Not(ss)
	case policy.OpLike:
		arg, ok := s.Argument.(string)
		if !ok {
			return nil, fmt.Errorf(`building "%s" statement: argument is not a string`, s.Operator)
		}
		return policy.Like(s.Selector, arg)
	default:
		return nil, fmt.Errorf("unknown statement: %s", s.Operator)
	}
}

func BuildStatements(statements ...Statement) ([]ucan.Statement, error) {
	stmts := make([]ucan.Statement, 0, len(statements))
	for i, s := range statements {
		ss, err := BuildStatement(s)
		if err != nil {
			return nil, fmt.Errorf("building statement %d: %w", i, err)
		}
		stmts = append(stmts, ss)
	}
	return stmts, nil
}

func Equal(sel string, value any) Statement {
	return Statement{
		Operator: policy.OpEqual,
		Selector: sel,
		Argument: value,
	}
}

func NotEqual(sel string, value any) Statement {
	return Statement{
		Operator: policy.OpNotEqual,
		Selector: sel,
		Argument: value,
	}
}

func GreaterThan(sel string, value any) Statement {
	return Statement{
		Operator: policy.OpGreaterThan,
		Selector: sel,
		Argument: value,
	}
}

func GreaterThanOrEqual(sel string, value any) Statement {
	return Statement{
		Operator: policy.OpGreaterThanOrEqual,
		Selector: sel,
		Argument: value,
	}
}

func LessThan(sel string, value any) Statement {
	return Statement{
		Operator: policy.OpLessThan,
		Selector: sel,
		Argument: value,
	}
}

func LessThanOrEqual(sel string, value any) Statement {
	return Statement{
		Operator: policy.OpLessThanOrEqual,
		Selector: sel,
		Argument: value,
	}
}

func Not(stmt Statement) Statement {
	return Statement{
		Operator: policy.OpNot,
		Argument: stmt,
	}
}

func And(stmts ...Statement) Statement {
	return Statement{
		Operator: policy.OpAnd,
		Argument: stmts,
	}
}

func Or(stmts ...Statement) Statement {
	return Statement{
		Operator: policy.OpOr,
		Argument: stmts,
	}
}

func Like(sel string, pattern string) Statement {
	return Statement{
		Operator: policy.OpLike,
		Selector: sel,
		Argument: pattern,
	}
}

func All(sel string, stmts ...Statement) Statement {
	return Statement{
		Operator: policy.OpAll,
		Selector: sel,
		Argument: stmts,
	}
}

func Any(sel string, stmts ...Statement) Statement {
	return Statement{
		Operator: policy.OpAny,
		Selector: sel,
		Argument: stmts,
	}
}
