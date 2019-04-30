package isaac

import (
	"sort"
)

func canCountVoting(total, threshold uint, yes, nop, exp int) bool {
	if threshold > total {
		return false
	}

	to := int(total)
	count := yes + nop + exp
	if count >= to {
		return true
	}

	th := int(threshold)

	margin := to - th

	// check majority
	if yes >= th || yes > margin {
		return true
	}

	if nop >= th || nop > margin {
		return true
	}

	if exp >= th || exp > margin {
		return true
	}

	// draw
	var voted = []int{yes, nop, exp}
	sort.Ints(voted)

	major := voted[len(voted)-1]

	if major+to-count < th {
		return true
	}

	return false
}

func majority(total, threshold uint, yes, nop, exp int) VoteResult {
	if !canCountVoting(total, threshold, yes, nop, exp) {
		return VoteResultNotYet // not yet
	}

	to := int(total)
	count := yes + nop + exp
	if count > to {
		return VoteResultDRAW
	}

	th := int(threshold)

	if nop >= th {
		return VoteResultNOP
	}

	if yes >= th {
		return VoteResultYES
	}

	if exp >= th {
		return VoteResultEXPIRE
	}

	return VoteResultDRAW
}
