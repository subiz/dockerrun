package main

import (
	"fmt"
)

func toString(s interface{}) string {
	if s == nil {
		return ""
	}
	switch v := s.(type) {
	case []byte:
		b := string(v)
		return b
	case string:
		return v
	case fmt.Stringer:
		str := v.String()
		return str
	default:
		str := fmt.Sprintf("%v", v)
		return str
	}
}
