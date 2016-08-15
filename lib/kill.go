package lib

import (
	"github.com/jimenez/mesoscon-demo/lib/mesosproto"
	"github.com/jimenez/mesoscon-demo/lib/mesosproto/schedulerproto"
)

func (lib *DemoLib) KillTask(taskId string) error {
	call := &schedulerproto.Call{
		FrameworkId: lib.frameworkID,
		Type:        schedulerproto.Call_KILL.Enum(),
		Kill: &schedulerproto.Call_Kill{
			AgentId: lib.tasks[taskId],
			TaskId: &mesosproto.TaskID{
				Value: &taskId,
			},
		},
	}

	_, err := lib.send(call, 202)
	return err
}
