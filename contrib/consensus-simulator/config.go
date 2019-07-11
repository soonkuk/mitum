package main

import (
	"encoding/json"
	"time"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v2"
)

type NodesGlobalConfig struct {
	Global        NodesNodeConfig
	NumberOfNodes uint          `yaml:"number-of-nodes"`
	ExitAfter     time.Duration `yaml:"exit-after"`
	Nodes         map[string]NodesNodeConfig
}

func newNodesConfigFromBytes(b []byte) (NodesGlobalConfig, error) {
	var gc NodesGlobalConfig

	if err := yaml.Unmarshal(b, &gc); err != nil {
		return NodesGlobalConfig{}, err
	}

	if err := gc.IsValid(); err != nil {
		return gc, err
	}

	nodes := map[string]NodesNodeConfig{}
	for name, c := range gc.Nodes {
		n := c.merge(gc.Global)

		if err := n.IsValid(); err != nil {
			return NodesGlobalConfig{}, err
		}

		nodes[name] = n
	}
	gc.Nodes = nodes

	return gc, nil
}

func (ng NodesGlobalConfig) String() string {
	b, _ := json.Marshal(ng)
	return string(b)
}

func (ng NodesGlobalConfig) IsValid() error {
	if err := ng.Global.IsValid(); err != nil {
		return err
	} else if ng.Global.Policy.Empty() {
		return xerrors.Errorf("empty global policy")
	} else if ng.Global.Block.Empty() {
		return xerrors.Errorf("empty global block")
	}

	if ng.NumberOfNodes < 1 {
		return xerrors.Errorf("NumberOfNodes should be greater than 0; NumberOfNodes=%q", ng.NumberOfNodes)
	}

	return nil
}

func (ng NodesGlobalConfig) Node(name string) NodesNodeConfig {
	c, found := ng.Nodes[name]
	if !found {
		return ng.Global
	}

	return c
}

type NodesBlockConfig struct {
	StartHeight uint64 `yaml:"start-height"`
	StartRound  uint64 `yaml:"start-round"`
}

func (nb NodesBlockConfig) Empty() bool {
	if nb.StartHeight > 0 {
		return false
	}

	return true
}

func (nb NodesBlockConfig) IsValid() error {
	if nb.Empty() {
		return nil
	}

	if nb.StartHeight < 1 {
		return xerrors.Errorf(
			"StartHeight should be greater than 0; StartHeight=%q",
			nb.StartHeight,
		)
	}

	return nil
}

func (nb NodesBlockConfig) merge(c NodesBlockConfig) NodesBlockConfig {
	if nb.StartHeight == 0 {
		nb.StartHeight = c.StartHeight
	}

	if c.StartRound > 0 {
		nb.StartRound = c.StartRound
	}

	return nb
}

type NodesPolicyConfig struct {
	TimeoutINITBallot        time.Duration `yaml:"timeout-init-ballot"`
	IntervalINITBallotOfJoin time.Duration `yaml:"interval-init-ballot-of-join"`
}

func (np NodesPolicyConfig) Empty() bool {
	if np.TimeoutINITBallot >= 0 {
		return false
	}

	if np.IntervalINITBallotOfJoin >= 0 {
		return false
	}

	return true
}

func (np NodesPolicyConfig) IsValid() error {
	if np.Empty() {
		return nil
	}

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

func (np NodesPolicyConfig) merge(c NodesPolicyConfig) NodesPolicyConfig {
	if np.TimeoutINITBallot <= 0 {
		np.TimeoutINITBallot = c.TimeoutINITBallot
	}

	if np.IntervalINITBallotOfJoin <= 0 {
		np.IntervalINITBallotOfJoin = c.IntervalINITBallotOfJoin
	}

	return np
}

type NodesModulesConfig struct {
	ProposalValidator map[string]interface{}
}

func (nm NodesModulesConfig) merge(c NodesModulesConfig) NodesModulesConfig {
	if _, ok := nm.ProposalValidator["name"]; !ok {
		nm.ProposalValidator = c.ProposalValidator
	}

	return nm
}

type NodesNodeConfig struct {
	Policy  NodesPolicyConfig
	Block   NodesBlockConfig
	Modules NodesModulesConfig
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

func (nn *NodesNodeConfig) UnmarshalYAML(um func(interface{}) error) error {
	var n struct {
		Policy NodesPolicyConfig
		Block  NodesBlockConfig
	}
	if err := um(&n); err != nil {
		return err
	}

	nn.Policy = n.Policy
	nn.Block = n.Block

	d := map[string]interface{}{}
	if err := um(d); err != nil {
		return err
	}

	log.Debug("trying to parse modules", "d", spew.Sdump(d))

	ms, ok := d["modules"]
	if !ok {
		log.Debug("empty modules")
		return nil
	}

	msm, ok := ms.(map[interface{}]interface{})
	if !ok {
		return xerrors.Errorf("modules should be map[string]interface{}; values=%q", ms)
	}

	mc := &NodesModulesConfig{}
	for k, v := range msm {
		kn, ok := k.(string)
		if !ok {
			return xerrors.Errorf("key should be string; key=%q", k)
		}

		if err := parseModule(mc, kn, v); err != nil {
			return err
		}
	}

	nn.Modules = *mc

	return nil
}

func (nn NodesNodeConfig) merge(c NodesNodeConfig) NodesNodeConfig {
	nn.Policy = nn.Policy.merge(c.Policy)
	nn.Block = nn.Block.merge(c.Block)
	nn.Modules = nn.Modules.merge(c.Modules)

	return nn
}

func parseModule(mc *NodesModulesConfig, name string, v interface{}) error {
	log.Debug("trying to parse module", "name", name)

	var err error
	switch name {
	case "ProposalValidator":
		err = parseModuleProposalValidator(mc, v)
	default:
		return xerrors.Errorf("unknown module found: name=%q", name)
	}

	if err != nil {
		return err
	}

	return nil
}

func parseModuleProposalValidator(mc *NodesModulesConfig, v interface{}) error {
	m, ok := v.(map[interface{}]interface{})
	if !ok {
		return xerrors.Errorf("module should be map[string]interface{}; values=%q", v)
	}

	n, ok := m["name"]
	if !ok {
		return xerrors.Errorf("empty name")
	}

	name, ok := n.(string)
	if !ok {
		return xerrors.Errorf("name should be string; name=%q", n)
	}

	mc.ProposalValidator = map[string]interface{}{
		"name": name,
	}

	var err error
	switch name {
	case "isaac.TestProposalValidator":
		err = parseModuleProposalValidatorTestProposalValidator(mc, m)
	}

	if err != nil {
		return err
	}

	return nil
}

func parseModuleProposalValidatorTestProposalValidator(mc *NodesModulesConfig, m map[interface{}]interface{}) error {
	var duration time.Duration
	d, ok := m["duration"]
	if !ok {
		return nil
	}

	if ds, ok := d.(string); !ok {
		return xerrors.Errorf("duration should be string; %q", d)
	} else if p, err := time.ParseDuration(ds); err != nil {
		return xerrors.Errorf("invalid duration; %w", err)
	} else {
		duration = p
	}

	mc.ProposalValidator["duration"] = duration

	return nil
}
