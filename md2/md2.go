package md2

import "strings"

var replTable = map[string]string{
	"_": "\\_",
	"*": "\\*",
	"[": "\\[",
	"]": "\\]",
	"(": "\\(",
	")": "\\)",
	"~": "\\~",
	"`": "\\`",
	">": "\\>",
	"#": "\\#",
	"+": "\\+",
	"-": "\\-",
	"=": "\\=",
	"|": "\\|",
	"{": "\\{",
	"}": "\\}",
	".": "\\.",
	"!": "\\!",
}

// EscapeText replaces unsafe md2 entries with slash-escaped values
//
// Generally useful for telegram bot api payloads, where escaping is required
// when using MarkdownV2 parse mode.
func EscapeText(value string) string {
	res := value
	for from, to := range replTable {
		res = strings.ReplaceAll(res, from, to)
	}

	return res
}
