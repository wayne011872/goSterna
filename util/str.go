package util

import (
	"bytes"
)

func StrAppend(strs ...string) string {
	var buffer bytes.Buffer
	for _, str := range strs {
		buffer.WriteString(str)
	}
	return buffer.String()
}