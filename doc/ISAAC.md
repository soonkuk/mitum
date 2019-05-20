# ISAAC Consensus Protocol

## Terminology

### Abbr

* `B(100)`: block `Height`
* `V(Y,N,E)`: vote
* `S(IT,SG,AC,AL)`: `Stage`
* `R(0)`: `Round`

### Proposer

In ISAAC, every `Round` has it's own `Proposer`. `Proposer` proposes the next block with the received transitions.


### Stage

* `INIT` -> (`PROPOSE`) -> `SIGN` -> `ACCEPT` -> `ALLCONFIRM`

### `INIT`

* `INIT` means node want to start new `Round`.
* `INIT` ballot has,
    - Who is `Proposer`
    - block `Height`
    - `Round`
* if node gets `INIT` ballots over threshold, start new `Round`.


### Node State

* `BOOT`: starting process
* `SYNC`: syncing blocks
* `JOIN`: joining consensus
* `CONSENSUS`: in consensus
* `STOPPED`: stopping process

#### basic transition

`BOOT` -> `JOIN` -> `SYNC` -> `CONSENSUS`

#### if consensus stucked or failed

`CONSENSUS` -> (failed) -> `JOIN` -> (`SYNC`) -> `CONSENSUS`


## Node State Transition

### `JOIN`

1. node starts and goes into `JOIN`
1. broadcasts `INIT`
1. receives the incoming ballots from others
1. follows areements of others


#### if the agreed block is unknown

1. move to `SYNC`


### `SYNC`

1. check the state of the other nodes
1. request the latest block to others
1. find the major block
1. request the missing blocks from other nodes
1. the received block should be over threshold
1. after syncing, move to `JOIN`

* node should return the block data and it's proof also
* if received blocks from others in same `Height` are different, validate them with their proof.
* proof is the seals of propose and ballots of each nodes.


## consensus stuck

### under `SIGN` or `ACCEPT` stage, no more ballot

* in `SIGN` or `ACCEPT`until waiting for given timeout, no more ballots are received

1. **timeouted**
1. request the latest ballot to others
1. wait for the given time
1. if consensus reaches with the received ballots from others, done.
1. **timeouted**
1. send `INIT` ballot to start new `Round`
    - current block `Height`
    - `Round` + 1


### in `INIT` stucked

1. wait for new `INIT`
1. **timeouted**
1. request the latest ballot to others
1. if consensus reaches with the received ballots from others, done.
1. **timeouted**
1. keep waiting, ...

* this means, node waits `INIT` ballot, but does not receive enough ballots.

### nodes have different block `Height` and `Round` of `INIT` ballot

* move to `SYNC`

