package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "print config",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 { // load config file
			log.Debug("trying to load config", "file", args[0])
			b, err := ioutil.ReadFile(args[0])
			if err != nil {
				cmd.Println("Error:", err.Error())
				os.Exit(1)
			}

			gc, err := newNodesConfigFromBytes(b)
			if err != nil {
				cmd.Println("Error:", err.Error())
				os.Exit(1)
			}

			globalConfig = gc
			log.Debug("config loaded", "config", globalConfig)

			g, err := json.MarshalIndent(globalConfig, "", "  ")
			if err != nil {
				cmd.Println("Error:", err.Error())
				os.Exit(1)
			}

			fmt.Println(strings.Repeat("=", 100))
			fmt.Println(string(bytes.TrimSpace(g)))
			fmt.Println(strings.Repeat("=", 100))
			os.Exit(0)
		}

		b, err := yaml.Marshal(NodesGlobalConfig{})
		if err != nil {
			cmd.Println("Error:", err.Error())
			os.Exit(1)
		}

		fmt.Println(strings.Repeat("=", 100))
		fmt.Println(string(bytes.TrimSpace(b)))
		fmt.Println(strings.Repeat("=", 100))
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
