package common

import (
	"bytes"
	"encoding/json"
)

func PrintJSON(v interface{}, indent bool, escapeHTML bool) string {
	var b []byte
	switch v.(type) {
	case []byte:
		b = v.([]byte)
	default:
		s, err := json.Marshal(v)
		if err != nil {
			return "<wrong>"
		}
		b = s
	}

	buffer := &bytes.Buffer{}
	e := json.NewEncoder(buffer)
	e.SetEscapeHTML(escapeHTML)
	if indent {
		e.SetIndent("", "  ")
	}

	err := e.Encode(json.RawMessage(b))
	if err != nil {
		return ""
	}

	return buffer.String()
}
