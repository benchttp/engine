package tests

import (
	"github.com/benchttp/engine/runner/internal/metrics"
)

// Predicate represents a comparison operator.
type Predicate string

const (
	EQ  Predicate = "EQ"
	NEQ Predicate = "NEQ"
	GT  Predicate = "GT"
	GTE Predicate = "GTE"
	LT  Predicate = "LT"
	LTE Predicate = "LTE"
)

// Validate returns ErrUnknownPredicate if p is not a know Predicate, else nil.
func (p Predicate) Validate() error {
	if _, ok := predicateSymbols[p]; !ok {
		return errWithDetails(ErrUnknownPredicate, p)
	}
	return nil
}

func (p Predicate) match(comparisonResult metrics.ComparisonResult) bool {
	sup := comparisonResult == metrics.SUP
	inf := comparisonResult == metrics.INF

	switch p {
	case EQ:
		return !sup && !inf
	case NEQ:
		return sup || inf
	case GT:
		return sup
	case GTE:
		return !inf
	case LT:
		return inf
	case LTE:
		return !sup
	default:
		panic("tests: unknown predicate: " + string(p))
	}
}

var predicateSymbols = map[Predicate]string{
	EQ:  "==",
	NEQ: "!=",
	GT:  ">",
	GTE: ">=",
	LT:  "<",
	LTE: "<=",
}

func (p Predicate) symbol() string {
	s, ok := predicateSymbols[p]
	if !ok {
		return "unknown predicate"
	}
	return s
}
