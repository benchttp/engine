package metrics

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
		panic("metrics.Aggregate.MetricOf: unknown Source: " + src)
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
