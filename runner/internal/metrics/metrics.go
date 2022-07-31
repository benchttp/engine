package metrics

import "reflect"

type Value interface{}

type Source string

const (
	ResponseTimeAvg     Source = "AVG"
	ResponseTimeMin     Source = "MIN"
	ResponseTimeMax     Source = "MAX"
	RequestFailCount    Source = "FAILURE_COUNT"
	RequestSuccessCount Source = "SUCCESS_COUNT"
	RequestCount        Source = "TOTAL_COUNT"
)

type Type uint8

const (
	TypeInt      Type = Type(reflect.Int)
	TypeDuration Type = 12
)

func (src Source) Type() Type {
	switch src {
	case ResponseTimeAvg, ResponseTimeMin, ResponseTimeMax:
		return TypeDuration
	case RequestFailCount, RequestSuccessCount, RequestCount:
		return TypeInt
	}
	panic(badSource(src))
}

func (agg Aggregate) MetricOf(src Source) Metric {
	var v interface{}
	switch src {
	case ResponseTimeAvg:
		v = agg.Avg
	case ResponseTimeMin:
		v = agg.Min
	case ResponseTimeMax:
		v = agg.Max
	case RequestFailCount:
		v = agg.FailureCount
	case RequestSuccessCount:
		v = agg.SuccessCount
	case RequestCount:
		v = agg.TotalCount
	default:
		panic(badSource(src))
	}
	return Metric{Source: src, Value: v}
}

type Metric struct {
	Source Source
	Value  Value
}

func (m Metric) Compare(to Metric) ComparisonResult {
	return compareMetrics(m, to)
}

func badSource(src Source) string {
	return "metrics: unknown Source: " + string(src)
}
