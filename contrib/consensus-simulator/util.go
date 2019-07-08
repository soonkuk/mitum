package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func printError(cmd *cobra.Command, err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n\n", err.Error())
	_ = cmd.Help()
}

func printFlags(cmd *cobra.Command, format string) interface{} {
	switch format {
	case "json":
		return printFlagsJSON(cmd)
	default:
		return printFlagsTerminal(cmd)
	}
}

func printFlagsJSON(cmd *cobra.Command) json.RawMessage {
	out := map[string]interface{}{}

	cmd.Flags().VisitAll(func(pf *pflag.Flag) {
		if pf.Name == "help" {
			return
		}

		out[fmt.Sprintf("--%s", pf.Name)] = map[string]interface{}{
			"default": fmt.Sprintf("%v", pf.DefValue),
			"value":   fmt.Sprintf("%v", pf.Value),
		}
	})

	b, _ := json.Marshal(out)

	return b
}

func printFlagsTerminal(cmd *cobra.Command) string {
	var b bytes.Buffer

	var flags []string
	cmd.Flags().VisitAll(func(pf *pflag.Flag) {
		if pf.Name == "help" {
			return
		}

		flags = append(flags, fmt.Sprintf("--%s=%v (default: %v)", pf.Name, pf.DefValue, pf.Value))
	})

	fmt.Fprintf(&b, strings.Join(flags, ", "))
	return b.String()
}
