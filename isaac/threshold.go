package isaac

import (
	"encoding/json"
	"math"
	"sync"

	"github.com/rs/zerolog"
	"golang.org/x/xerrors"
)

type Threshold struct {
	sync.RWMutex
	base      [3]uint // [2]uint{total, threshold, percent}
	threshold *sync.Map
}

func NewThreshold(baseTotal uint, basePercent float64) (*Threshold, error) {
	th, err := calculateThreshold(baseTotal, basePercent)
	if err != nil {
		return nil, err
	}

	return &Threshold{
		base:      [3]uint{baseTotal, th, uint(basePercent * 100)},
		threshold: &sync.Map{},
	}, nil
}

func (tr *Threshold) Get(stage Stage) (uint, uint) {
	if i, found := tr.threshold.Load(stage); found {
		t := i.([3]uint)
		return t[0], t[1]
	}

	tr.RLock()
	defer tr.RUnlock()

	return tr.base[0], tr.base[1]
}

func (tr *Threshold) SetBase(baseTotal uint, basePercent float64) error {
	th, err := calculateThreshold(baseTotal, basePercent)
	if err != nil {
		return err
	}

	tr.Lock()
	defer tr.Unlock()

	tr.base = [3]uint{baseTotal, th, uint(basePercent * 100)}

	return nil
}

func (tr *Threshold) Set(stage Stage, total uint, percent float64) error {
	th, err := calculateThreshold(total, percent)
	if err != nil {
		return err
	}

	tr.threshold.Store(stage, [3]uint{total, th, uint(percent * 100)})

	return nil
}

func (tr *Threshold) MarshalJSON() ([]byte, error) {
	tr.RLock()
	defer tr.RUnlock()

	thh := map[string]interface{}{}
	tr.threshold.Range(func(k, v interface{}) bool {
		thh[k.(Stage).String()] = flattenThreshold(v.([3]uint))

		return true
	})

	return json.Marshal(map[string]interface{}{
		"base":      flattenThreshold(tr.base),
		"threshold": thh,
	})
}

func (tr *Threshold) MarshalZerologObject(e *zerolog.Event) {
	tr.RLock()
	defer tr.RUnlock()

	thh := zerolog.Dict()
	tr.threshold.Range(func(k, v interface{}) bool {
		t := v.([3]uint)
		thh.Uints(k.(Stage).String(), t[:])

		return true
	})

	e.Uints("base", tr.base[:])
	e.Dict("threshold", thh)
}

func (tr *Threshold) String() string {
	b, _ := json.Marshal(tr) // nolint
	return string(b)
}

func calculateThreshold(total uint, percent float64) (uint, error) {
	if percent > 100 {
		return 0, xerrors.Errorf("basePercent is over 100; %v", percent)
	}

	return uint(math.Ceil(float64(total) * (percent / 100))), nil
}

func flattenThreshold(a [3]uint) [3]interface{} {
	return [3]interface{}{
		a[0],
		a[1],
		float64(a[2]) / 100,
	}
}
