package lib

import "github.com/jimenez/mesoscon-demo/lib/mesosproto"

func (lib *DemoLib) Acknowledge(taskId *mesosproto.TaskID, AgentId *mesosproto.AgentID, UUID []byte) error {
	call := &mesosproto.Call{
		FrameworkId: lib.frameworkID,
		Type:        mesosproto.Call_ACKNOWLEDGE.Enum(),
		Acknowledge: &mesosproto.Call_Acknowledge{
			AgentId: AgentId,
			TaskId:  taskId,
			Uuid:    UUID,
		},
	}

	_, err := lib.send(call, 202)
	return err
}
