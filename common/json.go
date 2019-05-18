package common

import (
	"bytes"
	"encoding/json"
)

func PrintJSON(v interface{}, indent bool, escapeHTML bool) string {
	b, err := EncodeJSON(v, indent, escapeHTML)
	if err != nil {
		return ""
	}
	return string(b)
}

func EncodeJSON(v interface{}, indent, escapeHTML bool) ([]byte, error) {
	buffer := &bytes.Buffer{}
	e := json.NewEncoder(buffer)
	if indent {
		e.SetIndent("", "  ")
	}
	e.SetEscapeHTML(escapeHTML)

	err := e.Encode(v)
	return bytes.TrimRight(buffer.Bytes(), "\n"), err
}
