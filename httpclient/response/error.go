package response

type errorResponse struct {
	Errors []string `json:"errors"`
	Type   string   `json:"type"`
}

func ErrorClient(e []error) Response {
	return newResponse(errorResponse{
		Errors: errorsToString(e),
		Type:   "client",
	})
}

func ErrorServer(e []error) Response {
	return newResponse(errorResponse{
		Errors: errorsToString(e),
		Type:   "server",
	})
}

func errorsToString(e []error) []string {
	s := make([]string, len(e))
	for i, err := range e {
		s[i] = err.Error()
	}
	return s
}
