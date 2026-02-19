package paging

// Normalize clamps limit/offset to safe defaults.
func Normalize(limit int, offset int) (int, int) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
