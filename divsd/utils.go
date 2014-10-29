package divsd

import "strings"

// skip the first N fields in a string
func skipFields(msg string, sep string, num int) string {
	res := msg[:]
	for i := 0 ; i < num; i++ {
		res = res[strings.Index(res, sep) + 1:]
	}
	return res
}

