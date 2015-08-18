package lib

import "github.com/jimenez/mesoscon-demo/lib/mesosproto"

func (lib *DemoLib) KillTask(taskId string) error {
	call := &mesosproto.Call{
		FrameworkId: lib.frameworkID,
		Type:        mesosproto.Call_KILL.Enum(),
		Kill: &mesosproto.Call_Kill{
			AgentId: lib.tasks[taskId],
			TaskId: &mesosproto.TaskID{
				Value: &taskId,
			},
		},
	}

	_, err := lib.send(call, 202)
	return err
}
