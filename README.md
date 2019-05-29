# mitum

Ready to go to winter.

[![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/spikeekips/mitum) 
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fspikeekips%2Fmitum.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fspikeekips%2Fmitum?ref=badge_shield)
[![Go Report Card](https://goreportcard.com/badge/github.com/spikeekips/mitum)](https://goreportcard.com/report/github.com/spikeekips/mitum)
[![](https://tokei.rs/b1/github/spikeekips/mitum?category=lines)](https://github.com/spikeekips/mitum)


# Why I rewrite SEBAK

- New design of data
- Everything is *state*
- Clean and simple code
- Testing new consensus; replacable consensus
- Better networking layer
- Better data storage layer

# Differences With SEBAK

* balance and block height will use big integer instead of `uint64`
* ISAAC will be simpler and easier to understand
* no `EXP` ballot
* `INIT` ballot newly added
* every message is `Seal`ed
* stage transition will be blocked
* no client's API
* state db for account
* operations can be programed by user friendly script for account data

# New Feature

- Account state
- Multisig
- HD Wallet

# Networking

- Only for communication between nodes
- Efficient serialization

# Data

- State: IAVL(https://godoc.org/github.com/tendermint/iavl)

# Storage

- Fast read and write: leveldb
- Remote access thru gRPC

# Etc

- New amount type: math/big


## Test

```
$ go test -race -tags test ./... -v
```

```
$ golangci-lint run
```
