package response

import (
	"time"

	"github.com/benchttp/engine/runner"
)

func Report(rep *runner.Report) Response {
	return newResponse(reportResponse{
		Metadata: metadataResponse{
			FinishedAt:    rep.Metadata.FinishedAt,
			TotalDuration: rep.Metadata.TotalDuration,
		},
		Tests: testsResponse{
			Pass:    rep.Tests.Pass,
			Results: toTestResultsResponse(rep.Tests.Results),
		},
		Metrics: metricsResponse{
			ResponseTimes:           toTimeStatsResponse(rep.Metrics.ResponseTimes),
			StatusCodesDistribution: rep.Metrics.StatusCodesDistribution,
			RequestEventTimes:       toRequestEventTimesResponse(rep.Metrics.RequestEventTimes),
			Records:                 toRecordsResponse(rep.Metrics.Records),
			RequestFailures:         toRequestFailuresResponse(rep.Metrics.RequestFailures),
			RequestCount:            rep.Metrics.RequestCount(),
			RequestSuccessCount:     rep.Metrics.RequestSuccessCount(),
			RequestFailureCount:     rep.Metrics.RequestFailureCount(),
		},
	})
}

type reportResponse struct {
	Metadata metadataResponse `json:"metadata"`
	Metrics  metricsResponse  `json:"metrics"`
	Tests    testsResponse    `json:"tests"`
}

type metadataResponse struct {
	FinishedAt    time.Time     `json:"finishedAt"`
	TotalDuration time.Duration `json:"totalDuration"`
}

type testsResponse struct {
	Pass    bool                 `json:"pass"`
	Results []testResultResponse `json:"results"`
}

type testResultResponse struct {
	Pass  bool              `json:"pass"`
	Got   interface{}       `json:"got"`
	Input testInputResponse `json:"input"`
}

type testInputResponse struct {
	Name      string      `json:"name"`
	Field     string      `json:"field"`
	Predicate string      `json:"predicate"`
	Target    interface{} `json:"target"`
}

type metricsResponse struct {
	ResponseTimes           timeStatsResponse            `json:"responseTimes"`
	StatusCodesDistribution map[int]int                  `json:"statusCodesDistribution"`
	RequestEventTimes       map[string]timeStatsResponse `json:"requestEventTimes"`
	Records                 []recordReponse              `json:"records"`
	RequestFailures         []requestFailureResponse     `json:"requestFailures"`
	RequestCount            int                          `json:"requestCount"`
	RequestSuccessCount     int                          `json:"requestSuccessCount"`
	RequestFailureCount     int                          `json:"requestFailureCount"`
}

type timeStatsResponse struct {
	Min       time.Duration   `json:"min"`
	Max       time.Duration   `json:"max"`
	Mean      time.Duration   `json:"mean"`
	Median    time.Duration   `json:"median"`
	StdDev    time.Duration   `json:"standardDeviation"`
	Quartiles []time.Duration `json:"quartiles"`
	Deciles   []time.Duration `json:"deciles"`
}

type recordReponse struct {
	ResponseTime time.Duration `json:"responseTime"`
}

type requestFailureResponse struct {
	Reason string `json:"reason"`
}

func toTestResultsResponse(testResults []runner.TestCaseResult) []testResultResponse {
	resp := make([]testResultResponse, len(testResults))
	for i, r := range testResults {
		resp[i] = testResultResponse{
			Pass: r.Pass,
			Got:  r.Got,
			Input: testInputResponse{
				Name:      r.Input.Name,
				Field:     string(r.Input.Field),
				Predicate: string(r.Input.Predicate),
				Target:    r.Input.Target,
			},
		}
	}
	return resp
}

func toTimeStatsResponse(stats runner.MetricsTimeStats) timeStatsResponse {
	return timeStatsResponse{
		Min:       stats.Min,
		Max:       stats.Max,
		Mean:      stats.Mean,
		Median:    stats.Median,
		StdDev:    stats.StdDev,
		Quartiles: stats.Quartiles,
		Deciles:   stats.Deciles,
	}
}

func toRequestEventTimesResponse(in map[string]runner.MetricsTimeStats) map[string]timeStatsResponse {
	resp := map[string]timeStatsResponse{}
	for k, v := range in {
		resp[k] = toTimeStatsResponse(v)
	}
	return resp
}

func toRecordsResponse(in []struct{ ResponseTime time.Duration }) []recordReponse {
	resp := make([]recordReponse, len(in))
	for i, v := range in {
		resp[i] = recordReponse{ResponseTime: v.ResponseTime}
	}
	return resp
}

func toRequestFailuresResponse(in []struct{ Reason string }) []requestFailureResponse {
	resp := make([]requestFailureResponse, len(in))
	for i, v := range in {
		resp[i] = requestFailureResponse{Reason: v.Reason}
	}
	return resp
}
