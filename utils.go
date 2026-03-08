package things3

import "time"

// comparePtrTimeCmp compares two *time.Time pointers for use with slices.SortFunc.
// nil values are sorted to the end.
func comparePtrTimeCmp(a, b *time.Time) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return 1
	}
	if b == nil {
		return -1
	}
	return a.Compare(*b)
}
