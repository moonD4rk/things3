package things3

import (
	"net/url"
	"strings"
	"time"
)

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

// encodeQuery encodes url.Values for Things URL scheme.
// Things expects %20 for spaces, not + (which is standard form encoding).
// This is safe because original + characters are encoded as %2B by url.Values.Encode().
func encodeQuery(query url.Values) string {
	return strings.ReplaceAll(query.Encode(), "+", "%20")
}
