package isaac

import (
	"sort"
)

func canCountVoting(total, threshold uint, yes, nop int) bool {
	if threshold > total {
		threshold = total
	}

	to := int(total)
	count := yes + nop
	if count >= to {
		return true
	}

	th := int(threshold)

	// check majority
	if yes >= th {
		return true
	}

	if nop >= th {
		return true
	}

	// check draw
	var voted = []int{yes, nop}
	sort.Ints(voted)

	return voted[0] > to-th // min over margin
}

func majority(total, threshold uint, yes, nop int) VoteResult {
	if !canCountVoting(total, threshold, yes, nop) {
		return VoteResultNotYet // not yet
	}

	if threshold > total {
		threshold = total
	}

	th := int(threshold)

	if yes >= th {
		return VoteResultYES
	}

	if nop >= th {
		return VoteResultNOP
	}

	return VoteResultDRAW
}
