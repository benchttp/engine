package response

import (
	"encoding/json"
	"io"
)

type Response struct {
	payload interface{}
}

func (r Response) EncodeJSON(w io.Writer) error {
	return json.NewEncoder(w).Encode(r.payload)
}

func newResponse(payload interface{}) Response {
	return Response{payload: payload}
}
