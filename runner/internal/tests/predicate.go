package tests

import "fmt"

type Predicate string

const (
	EQ  Predicate = "EQ"
	NEQ Predicate = "NEQ"
	GT  Predicate = "GT"
	GTE Predicate = "GTE"
	LT  Predicate = "LT"
	LTE Predicate = "LTE"
)

func (p Predicate) Apply(left, right Value) SingleResult {
	pass := p.passFunc()(left, right)
	return SingleResult{
		Pass:    pass,
		Explain: p.explain(left, right, pass),
	}
}

func (p Predicate) passFunc() func(left, right Value) bool {
	return func(left, right Value) bool {
		switch p {
		case EQ:
			return left == right
		case NEQ:
			return left != right
		case GT:
			return left > right
		case GTE:
			return left >= right
		case LT:
			return left < right
		case LTE:
			return left <= right
		default:
			panic(fmt.Sprintf("%s: unknown predicate", p))
		}
	}
}

func (p Predicate) explain(metric, compared Value, pass bool) string {
	return fmt.Sprintf("want %s %d, got %d", p.Symbol(), compared, metric)
}

func (p Predicate) Symbol() string {
	switch p {
	case EQ:
		return "=="
	case NEQ:
		return "!="
	case GT:
		return ">"
	case GTE:
		return ">="
	case LT:
		return "<"
	case LTE:
		return "<="
	default:
		return "[unknown predicate]"
	}
}
