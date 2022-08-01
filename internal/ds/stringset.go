package ds

// Set is a collection of unique string values.
type StringSet map[string]struct{}

// Add adds v to StringSet and returns true, unless StringSet[v]
// already exists. In that case it is noop and returns false.
func (set StringSet) Add(v string) (ok bool) {
	if _, exists := set[v]; exists {
		return false
	}
	set[v] = struct{}{}
	return true
}
