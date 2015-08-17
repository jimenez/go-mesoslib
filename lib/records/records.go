package records

import (
	"bufio"
	"io"
	"net/textproto"
	"strconv"
)

type Decoder interface {
	Decode(v interface{}) error
}

type Unmarshaler func([]byte, interface{}) error

type decoder struct {
	in *textproto.Reader
	un Unmarshaler
}

var empty = []byte{}

func NewDecoder(in io.Reader, un Unmarshaler) Decoder {
	bio := bufio.NewReader(in)
	return &decoder{
		in: textproto.NewReader(bio),
		un: un,
	}
}

func (d *decoder) Decode(v interface{}) error {
	ll, err := d.in.ReadLine()
	if err != nil {
		return err
	}

	nbytes, err := strconv.ParseInt(ll, 10, 64)
	if err != nil {
		return err
	}

	// zero-length messages are allowed, should we return a secondary param
	// to indicate this condition?
	if nbytes == 0 {
		return d.un(empty, v)
	}

	// TODO(jdef) enforce max message size here

	buf := make([]byte, int(nbytes))
	_, err = io.ReadFull(d.in.R, buf)
	if err != nil {
		return err
	}
	return d.un(buf, v)
}
