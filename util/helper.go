package util

import (
	"strings"
)

// AddPrefixPerLine ...
func AddPrefixPerLine(text string, prefix string) string {
	sb := NewStringBuilder()
	defer sb.Release()

	for _, v := range strings.Split(text, "\n") {
		if !sb.IsEmpty() {
			sb.AppendByte('\n')
		}
		sb.AppendString(prefix, v)
	}
	return sb.String()
}

// ConcatString ...
func ConcatString(args ...string) string {
	sb := NewStringBuilder()
	defer sb.Release()
	for _, v := range args {
		sb.AppendString(v)
	}
	return sb.String()
}
