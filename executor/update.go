package executor

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/jimenez/go-mesoslib/mesosproto"
	"github.com/jimenez/go-mesoslib/mesosproto/executorproto"
)

// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func (lib *ExecutorLib) update(task *mesosproto.TaskInfo, state *mesosproto.TaskState) error {
	// uuid, err := newUUID()
	// if err != nil {
	// 	return err
	// }
	call := &executorproto.Call{
		Type:        executorproto.Call_UPDATE.Enum(),
		FrameworkId: lib.frameworkID,
		ExecutorId:  lib.executorID,
		Update: &executorproto.Call_Update{
			Status: &mesosproto.TaskStatus{
				TaskId: task.GetTaskId(),
				State:  state,
				Uuid:   []byte("test-abcd-ef-3455-454-001"),
			},
		},
	}
	_, err := lib.send(call, 202)
	return err
}
