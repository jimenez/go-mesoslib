package lib

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gogo/protobuf/proto"
	"github.com/jimenez/mesoscon-demo/lib/mesosproto"
)

func (lib *DemoLib) Acknowledge(taskId *mesosproto.TaskID, AgentId *mesosproto.AgentID, UUID []byte) error {
	call := mesosproto.Call{
		FrameworkId: lib.frameworkID,
		Type:        mesosproto.Call_ACKNOWLEDGE.Enum(),
		Acknowledge: &mesosproto.Call_Acknowledge{
			AgentId: AgentId,
			TaskId:  taskId,
			Uuid:    UUID,
		},
	}

	body, err := proto.Marshal(&call)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "http://"+lib.master+ENDPOINT, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 202 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s", body)
	}
	return nil
}
