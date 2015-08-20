package lib

import (
	"github.com/gogo/protobuf/proto"
	"github.com/jimenez/mesoscon-demo/lib/mesosproto"
)

func (lib *DemoLib) Checkpoint(taskId string) error {
	call := &mesosproto.Call{
		FrameworkId: lib.frameworkID,
		Type:        mesosproto.Call_CRIU.Enum(),
		Criu: &mesosproto.Call_Criu{
			Type:    mesosproto.Call_Criu_CHECKPOINT.Enum(),
			AgentId: lib.tasks[taskId],
			TaskId: &mesosproto.TaskID{
				Value: &taskId,
			},
			ContainerId: proto.String(lib.containers[taskId]),
		},
	}

	_, err := lib.send(call, 202)
	if err != nil {
		return err
	}
	return nil
}

func (lib *DemoLib) Restore(taskId string) error {
	call := &mesosproto.Call{
		FrameworkId: lib.frameworkID,

		Type: mesosproto.Call_CRIU.Enum(),
		Criu: &mesosproto.Call_Criu{
			Type:    mesosproto.Call_Criu_RESTORE.Enum(),
			AgentId: lib.tasks[taskId],
			TaskId: &mesosproto.TaskID{
				Value: &taskId,
			},
			ContainerId: proto.String(lib.containers[taskId]),
		},
	}

	_, err := lib.send(call, 202)
	if err != nil {
		return err
	}
	return nil
}
