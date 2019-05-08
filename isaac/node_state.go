package isaac

type NodeState uint

const (
	_ NodeState = iota
	NodeStateBooting
	NodeStateJoining
	NodeStateConsensus
	NodeStateSync
	NodeStateStopped
)

func (n NodeState) IsValid() error {
	switch n {
	case NodeStateBooting:
	case NodeStateJoining:
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
	case NodeStateJoining:
		return "joining"
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
