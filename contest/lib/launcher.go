package contestlib

import (
	"fmt"
	"net/url"

	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/launcher"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/cache"
	"github.com/spikeekips/mitum/util/logging"
	"golang.org/x/xerrors"
)

type Launcher struct {
	*logging.Logging
	*launcher.Launcher
	design *launcher.NodeDesign
}

func NewLauncherFromDesign(design *launcher.NodeDesign, version util.Version) (*Launcher, error) {
	nr := &Launcher{design: design}

	if bn, err := launcher.NewLauncher(design, version); err != nil {
		return nil, err
	} else {
		nr.Launcher = bn
	}

	nr.Logging = logging.NewLogging(func(c logging.Context) logging.Emitter {
		return c.Str("module", "contest-node-runner")
	})
	return nr, nil
}

func (nr *Launcher) SetLogger(l logging.Logger) logging.Logger {
	_ = nr.Launcher.SetLogger(l)
	_ = nr.Logging.SetLogger(l)

	return nr.Log()
}

func (nr *Launcher) Design() *launcher.NodeDesign {
	return nr.design
}

func (nr *Launcher) Initialize() error {
	if err := nr.attachStorage(); err != nil {
		return err
	}

	if err := nr.attachNetwork(); err != nil {
		return err
	}

	if err := nr.attachNodeChannel(); err != nil {
		return err
	}

	if err := nr.attachRemoteNodes(); err != nil {
		return err
	}

	if err := nr.attachSuffrage(); err != nil {
		return err
	}

	if err := nr.attachProposalProcessor(); err != nil {
		return err
	}

	return nr.Launcher.Initialize()
}

func (nr *Launcher) attachStorage() error {
	l := nr.Log().WithLogger(func(ctx logging.Context) logging.Emitter {
		return ctx.Str("target", "storage")
	})
	l.Debug().Msg("trying to attach")

	var ca cache.Cache
	if len(nr.design.Component.StorageCache()) > 0 {
		if c, err := cache.NewCacheFromURI(nr.design.Component.StorageCache()); err != nil {
			return err
		} else {
			ca = c
		}
	}

	if st, err := launcher.LoadStorage(nr.design.Storage, nr.Encoders(), ca); err != nil {
		return err
	} else {
		_ = nr.SetStorage(st)
	}

	l.Debug().Msg("attached")

	return nil
}

func (nr *Launcher) attachNetwork() error {
	l := nr.Log().WithLogger(func(ctx logging.Context) logging.Emitter {
		return ctx.Str("target", "network")
	})
	l.Debug().Msg("trying to attach")

	if qs, err := launcher.LoadNetworkServer(
		nr.design.Network.Bind().Host,
		nr.design.Network.PublishURL(),
		nr.Encoders(),
	); err != nil {
		return err
	} else {
		_ = nr.SetNetwork(qs).
			SetPublichURL(nr.design.Network.PublishURL().String())
	}

	l.Debug().Msg("attached")

	return nil
}

func (nr *Launcher) attachNodeChannel() error {
	l := nr.Log().WithLogger(func(ctx logging.Context) logging.Emitter {
		return ctx.Str("target", "node-channel")
	})
	l.Debug().Msg("trying to attach")

	nu := new(url.URL)
	*nu = *nr.design.Network.PublishURL()
	nu.Host = fmt.Sprintf("localhost:%s", nu.Port())

	if ch, err := launcher.LoadNodeChannel(nu, nr.Encoders()); err != nil {
		return err
	} else {
		if l, ok := ch.(logging.SetLogger); ok {
			_ = l.SetLogger(nr.Log())
		}

		_ = nr.SetNodeChannel(ch)
		_ = nr.Localstate().Node().SetChannel(ch)
	}

	l.Debug().Msg("attached")

	return nil
}

func (nr *Launcher) attachRemoteNodes() error {
	nodes := make([]network.Node, len(nr.design.Nodes))

	for i := range nr.design.Nodes {
		c := nr.design.Nodes[i]
		l := nr.Log().WithLogger(func(ctx logging.Context) logging.Emitter {
			return ctx.Str("target", "remote-nodes").Str("address", c.Address().String())
		})

		l.Debug().Str("url", c.NetworkURL().String()).Msg("trying to create remote node")

		n := isaac.NewRemoteNode(c.Address(), c.Publickey(), c.NetworkURL().String())
		if ch, err := launcher.LoadNodeChannel(c.NetworkURL(), nr.Encoders()); err != nil {
			return err
		} else {
			if l, ok := ch.(logging.SetLogger); ok {
				_ = l.SetLogger(nr.Log())
			}

			_ = n.SetChannel(ch)
		}
		l.Debug().Msg("created")

		nodes[i] = n
	}

	return nr.Localstate().Nodes().Add(nodes...)
}

func (nr *Launcher) attachSuffrage() error {
	l := nr.Log().WithLogger(func(ctx logging.Context) logging.Emitter {
		return ctx.Str("target", "suffrage")
	})
	l.Debug().Msg("trying to attach")

	if sf, err := nr.design.Component.Suffrage().New(nr.Localstate(), nr.Encoders()); err != nil {
		return xerrors.Errorf("failed to create new suffrage component: %w", err)
	} else {
		l.Debug().
			Str("type", nr.design.Component.Suffrage().Type).
			Interface("info", nr.design.Component.Suffrage().Info).
			Msg("suffrage loaded")

		_ = nr.SetSuffrage(sf)
	}

	l.Debug().Msg("attached")

	return nil
}

func (nr *Launcher) attachProposalProcessor() error {
	l := nr.Log().WithLogger(func(ctx logging.Context) logging.Emitter {
		return ctx.Str("target", "proposal-processor")
	})
	l.Debug().Msg("trying to attach")

	if pp, err := nr.design.Component.ProposalProcessor().New(nr.Localstate(), nr.Suffrage()); err != nil {
		return xerrors.Errorf("failed to create new proposal processor component: %w", err)
	} else {
		l.Debug().
			Str("type", nr.design.Component.ProposalProcessor().Type).
			Interface("info", nr.design.Component.ProposalProcessor().Info).
			Msg("proposal processor loaded")

		_ = nr.SetProposalProcessor(pp)
	}

	l.Debug().Msg("attached")

	return nil
}
