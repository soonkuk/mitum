package isaac

type NodeState uint

const (
	_ NodeState = iota
	NodeStateBooting
	NodeStateJoin
	NodeStateConsensus
	NodeStateSync
	NodeStateStopped
)

func (n NodeState) IsValid() error {
	switch n {
	case NodeStateBooting:
	case NodeStateJoin:
	case NodeStateConsensus:
	case NodeStateSync:
	case NodeStateStopped:
	default:
		return InvalidNodeStateError
	}

	return nil
}

func (n NodeState) String() string {
	switch n {
	case NodeStateBooting:
		return "booting"
	case NodeStateJoin:
		return "join"
	case NodeStateConsensus:
		return "consensus"
	case NodeStateSync:
		return "sync"
	case NodeStateStopped:
		return "stopped"
	default:
		return "<wrong node state>"
	}
}

func (n NodeState) CanVote() bool {
	switch n {
	case NodeStateJoin, NodeStateConsensus:
		return true
	}

	return false
}
