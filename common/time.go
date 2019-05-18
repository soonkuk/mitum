package common

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/beevik/ntp"
)

const (
	TIMEFORMAT_ISO8601 string = "2006-01-02T15:04:05.000000000Z07:00"
)

var (
	ZeroTime   Time = Time{Time: time.Time{}}
	timeSyncer *TimeSyncer
)

func FormatISO8601(t Time) string {
	return t.Time.Format(TIMEFORMAT_ISO8601)
}

func NowISO8601() string {
	return FormatISO8601(Now())
}

func ParseISO8601(s string) (Time, error) {
	t, err := time.Parse(TIMEFORMAT_ISO8601, s)
	if err != nil {
		return Time{}, err
	}

	return Time{Time: t}, err
}

type Time struct {
	time.Time
}

func (t Time) UTC() Time {
	return Time{Time: t.Time.UTC()}
}

func (t Time) String() string {
	return FormatISO8601(t)
}

func (t Time) MarshalBinary() ([]byte, error) {
	return Encode(t.String())
}

func (t *Time) UnmarshalBinary(b []byte) error {
	var s string
	if err := Decode(b, &s); err != nil {
		return err
	}

	n, err := ParseISO8601(s)
	if err != nil {
		return err
	}

	*t = n
	return nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(FormatISO8601(t))
}

func (t *Time) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	n, err := ParseISO8601(s)
	if err != nil {
		return err
	}

	*t = n

	return nil
}

func (t Time) Before(b Time) bool {
	return t.Time.Before(b.Time)
}

func (t Time) After(b Time) bool {
	return t.Time.After(b.Time)
}

func (t Time) Between(c Time, d time.Duration) bool {
	if d < 0 {
		d = d * -1
	}

	return t.Time.Before(c.Time.Add(d)) && t.Time.After(c.Time.Add(d*-1))
}

func (t Time) IsZero() bool {
	return t.Time.Equal(ZeroTime.Time)
}

func (t Time) Equal(a Time) bool {
	return t.Time.Equal(a.Time)
}

func (t Time) Sub(a Time) time.Duration {
	return t.Time.Sub(a.Time)
}

func (t Time) Add(a time.Duration) Time {
	return Time{Time: t.Time.Add(a)}
}

func Now() Time {
	if timeSyncer == nil {
		return Time{Time: time.Now()}
	}

	return Time{Time: time.Now().Add(timeSyncer.Offset())}
}

type TimeSyncer struct {
	sync.RWMutex
	*Logger
	server   string
	offset   time.Duration
	stopChan chan bool
	interval time.Duration
}

func NewTimeSyncer(server string, checkInterval time.Duration) (*TimeSyncer, error) {
	_, err := ntp.Query(server)
	if err != nil {
		return nil, err
	}

	return &TimeSyncer{
		Logger: NewLogger(
			log,
			"module", "time-syncer",
			"server", server,
			"interval", checkInterval,
		),
		server:   server,
		interval: checkInterval,
		stopChan: make(chan bool),
	}, nil
}

func SetTimeSyncer(syncer *TimeSyncer) {
	timeSyncer = syncer
	log.Debug("common.timeSyncer is set")
}

func (s *TimeSyncer) Start() error {
	s.Log().Debug("started")

	go s.schedule()

	return nil
}

func (s *TimeSyncer) Stop() error {
	s.Lock()
	defer s.Unlock()

	if s.stopChan != nil {
		s.stopChan <- true
		close(s.stopChan)
		s.stopChan = nil
	}

	s.Log().Debug("stopped")
	return nil
}

func (s *TimeSyncer) schedule() {
	ticker := time.NewTicker(s.interval)

end:
	for {
		select {
		case <-s.stopChan:
			ticker.Stop()
			break end
		case <-ticker.C:
			s.check()
		}
	}
}

func (s *TimeSyncer) Offset() time.Duration {
	return s.offset
}

func (s *TimeSyncer) check() {
	response, err := ntp.Query(s.server)
	if err != nil {
		s.Log().Error("failed to query", "error", err)
		return
	}

	if err := response.Validate(); err != nil {
		s.Log().Error("failed to validate response", "response", response, "error", err)
		return
	}

	s.Lock()
	defer s.Unlock()

	s.offset = response.ClockOffset

	s.Log().Debug("time checked", "response", response, "offset", s.offset)
}
