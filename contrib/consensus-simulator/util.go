package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spikeekips/mitum/hash"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/keypair"
	"github.com/spikeekips/mitum/node"
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

func newRandomBlock(height uint64, round uint64) isaac.Block {
	bk, _ := isaac.NewBlock(
		isaac.NewBlockHeight(height),
		isaac.Round(round),
		isaac.NewRandomProposalHash(),
	)

	return bk
}

func LogFileByNodeHandler(directory string, fmtr log15.Format, quiet bool) log15.Handler {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.MkdirAll(directory, os.ModePerm); err != nil {
			panic(err)
		}
	}

	logs := map[string]*os.File{}

	openFile := func(n string) io.Writer {
		f := filepath.Join(directory, fmt.Sprintf("%s.log", n))
		w, err := os.OpenFile(f, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		logs[n] = w
		return w
	}

	findNode := func(c []interface{}) string {
		for i := 0; i < len(c); i += 2 {
			if n, ok := c[i].(string); !ok {
				continue
			} else if n == "node" {
				return c[i+1].(string)
			}
		}

		return ""
	}

	_ = openFile("none")

	h := log15.FuncHandler(func(r *log15.Record) error {
		var w io.Writer
		n := findNode(r.Ctx)
		if len(n) < 1 {
			w = logs["none"]
		} else if f, found := logs[n]; found {
			w = f
		} else {
			w = openFile(n)
		}

		_, err := w.Write(fmtr.Format(r))
		return err
	})

	if !quiet {
		h = log15.MultiHandler(h, log15.StreamHandler(os.Stdout, fmtr))
	}

	return closingHandler{Handler: log15.LazyHandler(log15.SyncHandler(h)), files: logs}
}

type closingHandler struct {
	log15.Handler
	files map[string]*os.File
}

func (h *closingHandler) Close() error {
	for _, f := range h.files {
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}

func newPolicy() (isaac.Policy, error) {
	policy := isaac.NewTestPolicy()
	policy.TimeoutINITBallot = globalConfig.Global.Policy.TimeoutINITBallot
	policy.IntervalINITBallotOfJoin = globalConfig.Global.Policy.IntervalINITBallotOfJoin
	policy.BasePercent = 67

	threshold, err := isaac.NewThreshold(globalConfig.NumberOfNodes, policy.BasePercent)
	if err != nil {
		return isaac.Policy{}, err
	}
	policy.Threshold = threshold

	log.Debug("policy created", "policy", policy)

	return policy, nil
}

func newHome(n uint) node.Home {
	pk, _ := keypair.NewStellarPrivateKey()

	prefix := []byte(fmt.Sprintf("%d", n))

	var b [32]byte
	copy(b[:], prefix)

	h, _ := hash.NewHash(node.AddressHashHint, b[:])
	address := node.Address{Hash: h}

	return node.NewHome(address, pk)
}
