package strs

import "unicode/utf8"

func TakeFirstN(s string, n int, tail ...bool) string {
	if utf8.RuneCountInString(s) < n {
		return s
	}
	head := string([]rune(s)[:n])
	if len(tail) > 0 && tail[0] {
		return head + "..."
	}
	return head
}
