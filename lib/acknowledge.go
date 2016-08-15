package lib

import (
	"github.com/jimenez/mesoscon-demo/lib/mesosproto"
	"github.com/jimenez/mesoscon-demo/lib/mesosproto/schedulerproto"
)

func (lib *DemoLib) Acknowledge(taskId *mesosproto.TaskID, AgentId *mesosproto.AgentID, UUID []byte) error {
	call := &schedulerproto.Call{
		FrameworkId: lib.frameworkID,
		Type:        schedulerproto.Call_ACKNOWLEDGE.Enum(),
		Acknowledge: &schedulerproto.Call_Acknowledge{
			AgentId: AgentId,
			TaskId:  taskId,
			Uuid:    UUID,
		},
	}

	_, err := lib.send(call, 202)
	return err
}
