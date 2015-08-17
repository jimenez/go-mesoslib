package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/jimenez/mesoscon-demo/lib/mesosproto"
	"github.com/jimenez/mesoscon-demo/lib/records"
	"golang.org/x/net/context"
)

type EventChan <-chan *mesosproto.Event

type Subscription interface {
	Events() EventChan // closes when the subscription has terminated
	Close()
	Err() error // yields the error that caused the Events() channel to close
}

type subscription struct {
	body    io.ReadCloser
	events  chan *mesosproto.Event
	cancel  context.CancelFunc
	err     error
	errLock sync.Mutex
}

func (t *subscription) Events() EventChan {
	return EventChan(t.events)
}

func (t *subscription) Close() {
	t.cancel()
}

func (t *subscription) Err() error {
	t.errLock.Lock()
	defer t.errLock.Unlock()
	return t.err
}

func Subscribe(masterURI string, fi *mesosproto.FrameworkInfo, force bool) (Subscription, error) {
	call := mesosproto.Call{
		Type: mesosproto.Call_SUBSCRIBE.Enum(),
		Subscribe: &mesosproto.Call_Subscribe{
			FrameworkInfo: fi,
			Force: &force,
		},
	}

	body, err := proto.Marshal(&call)
	if err != nil {
		return nil, err
	}

	const EP_SCHEDULER = "/master/api/v1/scheduler"
	req, err := http.NewRequest("POST", "http://"+masterURI+EP_SCHEDULER, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Accept", "application/json")

	var t *subscription
	bg := context.Background()
	ctx, _ := context.WithTimeout(bg, 75*time.Second)
	err = httpDo(ctx, req, func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			return fmt.Errorf("%s", body)
		}
		c2, cancel := context.WithCancel(bg)
		t = &subscription{
			body:   resp.Body,
			events: make(chan *mesosproto.Event, 100),
			cancel: cancel,
		}
		go t.handleEvents(c2)
		return nil
	})
	return t, err
}

func httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
	// Run the HTTP request in a goroutine and pass the response to f.
	tr := &http.Transport{}
	client := &http.Client{Transport: tr}
	c := make(chan error, 1)
	go func() { c <- f(client.Do(req)) }()
	select {
	case <-ctx.Done():
		tr.CancelRequest(req)
		<-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func (t *subscription) setError(err error) {
	t.errLock.Lock()
	defer t.errLock.Unlock()
	t.err = err
}

func (t *subscription) handleEvents(ctx context.Context) {
	defer t.cancel()
	defer t.body.Close()
	defer close(t.events)

	var empty bool
	un := records.Unmarshaler(func(b []byte, v interface{}) error {
		if b == nil || len(b) == 0 {
			empty = true
			return nil
		}
		return json.Unmarshal(b, v)
	})
	errNoPulse := errors.New("failed to receive a pulse")
	dec := records.NewDecoder(t.body, un)
	timed := timedDecoder(dec, 15*time.Second, 5, errNoPulse)
	for {
		empty = false
		event := &mesosproto.Event{}
		if err := timed.Decode(event); err != nil {
			if err == io.EOF || err == errNoPulse {
				t.setError(err)
				return
			}
			// protocol error? log and attempt to continue
			// TODO(jdef) not all protocol errors are recoverable
			// TODO(jdef) log all of these
			log.Println("ERROR:", err)
			continue
		}
		if empty {
			// TODO(jdef) lame heartbeat?
			continue
		}
		select {
		case t.events <- event:
		case <-ctx.Done():
			// no async op to cancel, just abort
			t.setError(ctx.Err())
			return
		}
	}
}

// timedDecoder returns a Decorated decoder that generates the given error if no events
// are decoded for some number of sequential timeouts. The returned Decoder is not safe
// to share across goroutines.
// TODO(jdef) this probably isn't the right place for all of this logic (and it's not
// just monitoring the heartbeat messages, it's counting all of them..). Heartbeat monitoring
// has specific requirements. Get rid of this and implement something better elsewhere.
func timedDecoder(dec records.Decoder, dur time.Duration, timeouts int, err error) records.Decoder {
	var t *time.Timer
	return records.DecoderFunc(func(v interface{}) error {
		if t == nil {
			t = time.NewTimer(dur)
		} else {
			t.Reset(dur)
		}
		defer t.Stop()

		errCh := make(chan error, 1)
		go func() {
			// there's no way to abort this so someone else will have
			// to make sure that it dies (and it should if the response
			// body is closed)
			errCh <- dec.Decode(v)
		}()
		for x := 0; x < timeouts; x++ {
			select {
			case <-t.C:
				// check for a tie
				select {
				case e := <-errCh:
					return e
				default:
					// noop, continue
				}
			case e := <-errCh:
				return e
			}
		}
		return err
	})
}
