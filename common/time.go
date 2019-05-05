package common

import (
	"encoding/json"
	"time"
)

const (
	TIMEFORMAT_ISO8601 string = "2006-01-02T15:04:05.000000000Z07:00"
)

var (
	ZeroTime Time = Time{Time: time.Time{}}
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

func Now() Time {
	return Time{Time: time.Now()}
}
