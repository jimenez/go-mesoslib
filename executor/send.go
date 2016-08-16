package executor

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/protobuf/proto"
	"github.com/jimenez/go-mesoslib/mesosproto/executorproto"
)

func (lib *ExecutorLib) send(call *executorproto.Call, statusExpected int) (io.ReadCloser, error) {

	logrus.Infof("POST http://%s: \n%#v", lib.agent+ENDPOINT, call)

	spew.Dump(call)
	body, err := proto.Marshal(call)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "http://"+lib.agent+ENDPOINT, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != statusExpected {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%s", body)
	}

	return resp.Body, nil

}
