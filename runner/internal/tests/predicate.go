package tests

import (
	"github.com/benchttp/engine/runner/internal/metrics"
)

type Predicate string

const (
	EQ  Predicate = "EQ"
	NEQ Predicate = "NEQ"
	GT  Predicate = "GT"
	GTE Predicate = "GTE"
	LT  Predicate = "LT"
	LTE Predicate = "LTE"
)

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
