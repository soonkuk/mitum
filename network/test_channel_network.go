// +build test

package network

import (
	"context"
	"sync"

	"github.com/rs/zerolog"
	"github.com/spikeekips/mitum/common"
	"github.com/spikeekips/mitum/node"
	"github.com/spikeekips/mitum/seal"
	"golang.org/x/xerrors"
)

type ChannelNetworkSealHandler func(seal.Seal) (seal.Seal, error)

type ChannelNetwork struct {
	sync.RWMutex
	*common.ReaderDaemon
	home    node.Home
	chans   map[node.Address]*ChannelNetwork
	handler ChannelNetworkSealHandler
}

func NewChannelNetwork(home node.Home, handler ChannelNetworkSealHandler) *ChannelNetwork {
	cn := &ChannelNetwork{
		ReaderDaemon: common.NewReaderDaemon(false, 0, nil),
		home:         home,
		handler:      handler,
	}
	cn.ReaderDaemon.Logger = common.NewLogger(func(c zerolog.Context) zerolog.Context {
		return c.Str("module", "channel-suffrage-network")
	})
	cn.chans = map[node.Address]*ChannelNetwork{home.Address(): cn}

	return cn
}

func (cn *ChannelNetwork) Home() node.Home {
	return cn.home
}

func (cn *ChannelNetwork) AddMembers(chans ...*ChannelNetwork) *ChannelNetwork {
	cn.Lock()
	defer cn.Unlock()

	for _, ch := range chans {
		if ch.Home().Equal(cn.Home()) {
			continue
		}
		cn.chans[ch.Home().Address()] = ch
	}

	return cn
}

func (cn *ChannelNetwork) SetHandler(handler ChannelNetworkSealHandler) *ChannelNetwork {
	cn.handler = handler
	return cn
}

func (cn *ChannelNetwork) Broadcast(sl seal.Seal) error {
	cn.RLock()
	defer cn.RUnlock()

	var wg sync.WaitGroup
	wg.Add(len(cn.chans))

	for _, ch := range cn.chans {
		go func(ch *ChannelNetwork) {
			defer wg.Done()

			if ch.Write(sl) {
				cn.Log().Debug().Object("to", ch.Home().Address()).Object("seal", sl).Msg("sent seal")
			} else {
				cn.Log().Error().Object("to", ch.Home().Address()).Object("seal", sl).Msg("failed to send seal")
			}
		}(ch)
	}

	wg.Wait()

	return nil
}

func (cn *ChannelNetwork) Request(_ context.Context, n node.Address, sl seal.Seal) (seal.Seal, error) {
	cn.RLock()
	defer cn.RUnlock()

	ch, found := cn.chans[n]
	if !found {
		return nil, xerrors.Errorf("unknown node; node=%q", n)
	}

	if ch.handler == nil {
		return nil, xerrors.Errorf("node=%q handler not registered", n)
	}

	return ch.handler(sl)
}

func (cn *ChannelNetwork) RequestAll(ctx context.Context, sl seal.Seal) (map[node.Address]seal.Seal, error) {
	results := map[node.Address]seal.Seal{}

	cn.RLock()
	defer cn.RUnlock()

	for n := range cn.chans {
		r, err := cn.Request(ctx, n, sl)
		if err != nil {
			cn.Log().Error().Err(err).Object("target", n).Msg("failed to request")
		}
		results[n] = r
	}

	return results, nil
}
