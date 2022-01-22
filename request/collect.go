package request

func collect(c <-chan Record) []Record {
	rec := []Record{}

	for r := range c {
		rec = append(rec, r)
	}
	return rec
}
