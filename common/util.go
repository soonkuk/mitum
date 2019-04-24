package common

import (
	"bytes"
	"fmt"
	"strings"
)

func PrettyMap(m map[string]interface{}) string {
	b := new(bytes.Buffer)
	for k, v := range m {
		fmt.Fprintf(b, "%s=%v ", k, v)
	}

	return strings.TrimSpace(b.String())
}
