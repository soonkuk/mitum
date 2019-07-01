package isaac

/*
import (
	"sync"

	"github.com/spikeekips/mitum/common"
)

// BallotHandler manages voting process; it will block the vote one by one.
type BallotHandler struct {
	sync.RWMutex
	*common.Logger
	ballotbox *Ballotbox
	daemon    *common.ReaderDaemon
	receiver  chan interface{}
	threshold *Threshold
	homeState *HomeState
}

func NewBallotHandler(homeState *HomeState, threshold *Threshold) *BallotHandler {
	bc := &BallotHandler{
		Logger:    common.NewLogger(Log(), "module", "voting"),
		ballotbox: NewBallotbox(),
		threshold: threshold,
		homeState: homeState,
	}

	bc.daemon = common.NewReaderDaemon(true, bc.voteCallback)

	return vt
}

func (bc *BallotHandler) Start() error {
	bc.RLock()
	defer bc.RUnlock()

	if !bc.daemon.IsStopped() {
		return common.DaemonAleadyStartedError.Newf("BallotHandler is already running; daemon is still running")
	}

	if err := bc.daemon.Start(); err != nil {
		return err
	}

	return nil
}

func (bc *BallotHandler) Stop() error {
	bc.Lock()
	defer bc.Unlock()

	if err := bc.daemon.Stop(); err != nil {
		return err
	}

	return nil
}

func (bc *BallotHandler) Vote(ballot Ballot) {
	go func() {
		bc.daemon.Write(ballot)
	}()
}

func (bc *BallotHandler) voteCallback(v interface{}) error {
}
*/
