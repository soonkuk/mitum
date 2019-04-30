package common

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func PrintJSON(b []byte, indent bool, escapeHTML bool) string {
	buffer := &bytes.Buffer{}
	e := json.NewEncoder(buffer)
	e.SetEscapeHTML(escapeHTML)
	if indent {
		e.SetIndent("", "  ")
	}

	err := e.Encode(json.RawMessage(b))
	if err != nil {
		fmt.Println(">>>>>>.", err)
		return ""
	}

	return buffer.String()
}
