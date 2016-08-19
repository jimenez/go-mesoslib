package executor

import (
	"github.com/jimenez/go-mesoslib/mesosproto"
	"github.com/jimenez/go-mesoslib/mesosproto/executorproto"
)

func (lib *ExecutorLib) update(task *mesosproto.TaskInfo, state *mesosproto.TaskState) error {
	call := &executorproto.Call{
		Type:        executorproto.Call_UPDATE.Enum(),
		FrameworkId: lib.frameworkID,
		ExecutorId:  lib.executorID,
		Update: &executorproto.Call_Update{
			Status: &mesosproto.TaskStatus{
				TaskId: task.GetTaskId(),
				State:  state,
			},
		},
	}
	_, err := lib.send(call, 202)
	return err
}
