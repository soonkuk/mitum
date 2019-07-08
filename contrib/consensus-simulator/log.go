package main

import (
	"github.com/inconshreveable/log15"
	"github.com/spikeekips/mitum/common"
)

var log log15.Logger = log15.New("module", "main")

func init() {
	//handler, _ := LogHandler(LogFormatter("terminal"), "")
	handler, _ := common.LogHandler(common.LogFormatter("json"), "")
	handler = log15.CallerFileHandler(handler)
	log.SetHandler(log15.LvlFilterHandler(log15.LvlDebug, handler))
	//logger.SetHandler(log15.LvlFilterHandler(log15.LvlCrit, handler))
}
