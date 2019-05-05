package common

import (
	"encoding/json"

	"github.com/Masterminds/semver"
)

var (
	ZeroVersion Version = Version{}
)

type Version semver.Version

func NewVersion(s string) (Version, error) {
	v, err := semver.NewVersion(s)
	if err != nil {
		return Version{}, err
	}

	return Version(*v), nil
}

func MustParseVersion(s string) Version {
	v, err := NewVersion(s)
	if err != nil {
		panic(err)
	}
	return v
}

func (v Version) String() string {
	p := semver.Version(v)
	return (&p).String()
}

func (v Version) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v *Version) UnmarshalJSON(b []byte) error {
	var n string
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	} else if s, err := semver.NewVersion(n); err != nil {
		return err
	} else {
		*v = Version(*s)
	}

	return nil
}

func (v Version) MarshalBinary() ([]byte, error) {
	return json.Marshal(v)
}

func (v *Version) UnmarshalBinary(b []byte) error {
	return json.Unmarshal(b, v)
}

func (v Version) Equal(b Version) bool {
	a := semver.Version(v)
	c := semver.Version(b)

	return (&a).Equal(&c)
}
