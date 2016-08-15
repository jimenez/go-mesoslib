package executor

import (
	"github.com/jimenez/go-mesoslib/mesosproto"
	"github.com/jimenez/go-mesoslib/mesosproto/executorproto"
)

func (lib *ExecutorLib) update(task *mesosproto.TaskInfo, state *TaskState) {
	call := &executorproto.Call{
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
