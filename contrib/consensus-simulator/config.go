package main

import (
	"encoding/json"
	"time"

	"golang.org/x/xerrors"
	"gopkg.in/yaml.v2"
)

type NodesGlobalConfig struct {
	Global        NodesNodeConfig
	NumberOfNodes uint          `yaml:"number-of-nodes"`
	ExitAfter     time.Duration `yaml:"exit-after"`
}

func newNodesConfigFromBytes(b []byte) (NodesGlobalConfig, error) {
	var gc NodesGlobalConfig

	if err := yaml.Unmarshal(b, &gc); err != nil {
		return NodesGlobalConfig{}, err
	}

	if err := gc.IsValid(); err != nil {
		return gc, err
	}

	return gc, nil
}

func (ng NodesGlobalConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Global":        ng.Global,
		"NumberOfNodes": ng.NumberOfNodes,
		"ExitAfter":     ng.ExitAfter,
	})
}

func (ng NodesGlobalConfig) String() string {
	b, _ := json.Marshal(ng)
	return string(b)
}

func (ng NodesGlobalConfig) IsValid() error {
	if err := ng.Global.IsValid(); err != nil {
		return err
	}

	if ng.NumberOfNodes < 1 {
		return xerrors.Errorf("NumberOfNodes should be greater than 0; NumberOfNodes=%q", ng.NumberOfNodes)
	}

	return nil
}

type NodesNodeConfig struct {
	Policy NodesPolicyConfig
	Block  NodesBlockConfig
}

func (nn NodesNodeConfig) IsValid() error {
	if err := nn.Policy.IsValid(); err != nil {
		return err
	}

	if err := nn.Block.IsValid(); err != nil {
		return err
	}

	return nil
}

type NodesBlockConfig struct {
	StartHeight uint64 `yaml:"start-height"`
	StartRound  uint64 `yaml:"start-round"`
}

func (nb NodesBlockConfig) IsValid() error {
	if nb.StartHeight < 1 {
		return xerrors.Errorf(
			"StartHeight should be greater than 0; StartHeight=%q",
			nb.StartHeight,
		)
	}

	return nil
}

type NodesPolicyConfig struct {
	TimeoutINITBallot        time.Duration `yaml:"timeout-init-ballot"`
	IntervalINITBallotOfJoin time.Duration `yaml:"interval-init-ballot-of-join"`
}

func (np NodesPolicyConfig) IsValid() error {
	if np.TimeoutINITBallot < time.Millisecond {
		return xerrors.Errorf(
			"TimeoutINITBallot should be greater than 0; TimeoutINITBallot=%q",
			np.TimeoutINITBallot,
		)
	}

	if np.IntervalINITBallotOfJoin < time.Millisecond {
		return xerrors.Errorf(
			"IntervalINITBallotOfJoin should be greater than 0; IntervalINITBallotOfJoin=%q",
			np.IntervalINITBallotOfJoin,
		)
	}

	return nil
}
