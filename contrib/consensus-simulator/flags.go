package main

import (
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"golang.org/x/xerrors"
)

var (
	flagLogLevel      FlagLogLevel  = FlagLogLevel{lvl: log15.LvlDebug}
	flagLogFormat     FlagLogFormat = FlagLogFormat{f: "json"}
	FlagLogOut        string
	flagCPUProfile    string
	flagExitAfter     time.Duration
	flagNumberOfNodes uint = 3
	flagQuiet         bool
)

type FlagLogLevel struct {
	lvl log15.Lvl
}

func (f FlagLogLevel) String() string {
	return f.lvl.String()
}

func (f *FlagLogLevel) Set(v string) error {
	lvl, err := log15.LvlFromString(v)
	if err != nil {
		return err
	}

	f.lvl = lvl

	return nil
}

func (f FlagLogLevel) Type() string {
	return "log-level"
}

type FlagLogFormat struct {
	f string
}

func (f FlagLogFormat) String() string {
	return f.f
}

func (f *FlagLogFormat) Set(v string) error {
	s := strings.ToLower(v)
	switch s {
	case "json":
	case "terminal":
	default:
		return xerrors.Errorf("invalid log format: %q", v)
	}

	f.f = s

	return nil
}

func (f FlagLogFormat) Type() string {
	return "log-format"
}
