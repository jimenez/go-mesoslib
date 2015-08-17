package records

import (
	"bufio"
	"fmt"
	"io"
	"net/textproto"
	"strconv"
)

const DefaultMaxEventLength = uint64(1024 * 1024)

var (
	empty          = []byte{}
	MaxEventLength = DefaultMaxEventLength // should not be larger than what int will hold
)

type Decoder interface {
	Decode(v interface{}) error
}

// A ProtocolError describes a protocol violation such as an invalid `event-length`.
type ProtocolError string

func (p ProtocolError) Error() string {
	return string(p)
}

type DecoderFunc func(v interface{}) error

func (df DecoderFunc) Decode(v interface{}) error {
	return df(v)
}

type Unmarshaler func([]byte, interface{}) error

type decoder struct {
	in        *textproto.Reader
	un        Unmarshaler
	state     stateFn
	remaining int
	records   chan []byte
	buf       []byte
}

type stateFn func(d *decoder) (stateFn, error)

func NewDecoder(in io.Reader, un Unmarshaler) Decoder {
	bio := bufio.NewReader(in)
	return &decoder{
		in:      textproto.NewReader(bio),
		un:      un,
		records: make(chan []byte, 1),
		state:   headerState,
	}
}

func (d *decoder) Decode(v interface{}) (err error) {
stateLoop:
	for err == nil {
		if d.state, err = d.state(d); err == nil {
			select {
			case r := <-d.records:
				err = d.un(r, v)
				break stateLoop
			default:
			}
		}
	}
	return
}

func headerState(d *decoder) (stateFn, error) {
	ll, err := d.in.ReadLine()
	if err != nil {
		//TODO(jdef) check for EOF?
		return headerState, err
	}
	nbytes, err := strconv.ParseUint(ll, 10, 64)
	if err != nil {
		return failedState, ProtocolError(fmt.Sprintf("protocol violation, failed to parse event-length: %v", err))
	}
	// enforce max message size here
	if nbytes > MaxEventLength {
		// TODO(jdef) enter a "skip-bytes" state instead? if so, we may need to indicate
		// that the error is temporary
		return failedState, ProtocolError(fmt.Sprintf("protocol violation, event-length %d exceeds max allowed (%d)", nbytes, MaxEventLength))
	}
	if d.remaining = int(nbytes); d.remaining == 0 {
		d.buf = empty
	} else {
		// TODO(jdef) attempt to reuse existing buffer?
		d.buf = make([]byte, d.remaining)
	}
	return eventState, nil
}

func failedState(d *decoder) (stateFn, error) {
	return failedState, ProtocolError("decoder is in a failed state and will not recover")
}

func eventState(d *decoder) (stateFn, error) {
	var err error
	if d.remaining > 0 {
		off := len(d.buf) - d.remaining
		n, e := d.in.R.Read(d.buf[off:])
		if n > 0 {
			d.remaining -= n
		}
		err = e
	}
	if d.remaining == 0 {
		d.records <- d.buf
		// don't communicate an error in the same step that we generated a record,
		// let the next stateFn deal with any stream errors (e.g. EOF)
		return headerState, nil
	}
	return eventState, err
}
