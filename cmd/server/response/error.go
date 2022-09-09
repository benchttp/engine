package response

func Error(err error) Response {
	return newResponse(errorResponse{Error: err})
}

type errorResponse struct {
	Error error `json:"error"`
}
