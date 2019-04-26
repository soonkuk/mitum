package common

import (
	"net/url"
	"strconv"
)

type NetAddr struct {
	*url.URL
}

func NewNetAddr(a string) (NetAddr, error) {
	if len(a) < 1 {
		return NetAddr{}, InvalidNetAddrError
	}

	url, err := url.Parse(a)
	if err != nil {
		return NetAddr{}, InvalidNetAddrError.SetMessage(err.Error())
	}

	if len(url.Port()) > 0 {
		if _, err := strconv.ParseUint(url.Port(), 10, 64); err != nil {
			return NetAddr{}, InvalidNetAddrError.SetMessage(err.Error())
		}
	}

	return NetAddr{URL: url}, nil
}

func (n NetAddr) Network() string {
	return n.URL.Scheme
}

func (n NetAddr) String() string {
	return n.URL.String()
}

func (n NetAddr) MarshalBinary() ([]byte, error) {
	return n.URL.MarshalBinary()
}

func (n *NetAddr) UnmarshalBinary(b []byte) error {
	var u url.URL
	if err := u.UnmarshalBinary(b); err != nil {
		return InvalidNetAddrError.SetMessage(err.Error())
	}

	n.URL = &u

	return nil
}

func (n NetAddr) Equal(a NetAddr) bool {
	if n.URL.Scheme != a.URL.Scheme {
		return false
	}

	if n.URL.Hostname() != a.URL.Hostname() {
		return false
	}

	if n.URL.Port() != a.URL.Port() {
		return false
	}

	return true
}
