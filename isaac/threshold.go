package isaac

import (
	"sync"
)

type Threshold struct {
	sync.RWMutex
	base      [2]uint // [2]uint{total, threshold}
	threshold map[Stage][2]uint
}

func NewThreshold(baseTotal, baseThreshold uint) *Threshold {
	return &Threshold{
		base:      [2]uint{baseTotal, baseThreshold},
		threshold: map[Stage][2]uint{},
	}
}

func (tr *Threshold) Get(stage Stage) (uint, uint) {
	tr.RLock()
	defer tr.RUnlock()

	t, found := tr.threshold[stage]
	if found {
		return t[0], t[1]
	}

	return tr.base[0], tr.base[1]
}

func (tr *Threshold) SetBase(baseTotal, baseThreshold uint) *Threshold {
	tr.Lock()
	defer tr.Unlock()

	tr.base = [2]uint{baseTotal, baseThreshold}

	return tr
}

func (tr *Threshold) Set(stage Stage, total, threshold uint) *Threshold {
	tr.Lock()
	defer tr.Unlock()

	tr.threshold[stage] = [2]uint{total, threshold}

	return tr
}
