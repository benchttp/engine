package response

import (
	"encoding/json"
	"io"
)

type ErrorResponse struct {
	Error error `json:"error"`
}

func (resp ErrorResponse) EncodeJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(resp)
}

func Error(err error) ErrorResponse {
	return ErrorResponse{Error: err}
}
