package configio

import "github.com/benchttp/sdk/benchttp"

// UnmarshalJSONRunner parses the JSON-encoded data and stores the result
// in the benchttp.Runner pointed to by dst.
func UnmarshalJSONRunner(in []byte, dst *benchttp.Runner) error {
	repr := Representation{}
	if err := (JSONParser{}).Parse(in, &repr); err != nil {
		return err
	}
	return repr.Into(dst)
}
