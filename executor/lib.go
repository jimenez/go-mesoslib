package executor

import (
	"github.com/jimenez/go-mesoslib/mesosproto"
	"github.com/jimenez/go-mesoslib/mesosproto/executorproto"
)

const ENDPOINT = "/api/v1/executor"

type ExecutorLib struct {
	name             string
	agent            *mesosproto.AgentInfo
	frameworkID      *mesosproto.FrameworkID
	executorID       *mesosproto.ExecutorID
	tasks            map[string]*mesosproto.TaskInfo
	tasksUnAkowledge map[string]*mesosproto.TaskInfo
}

func New(name, frameworkId, executorId, hostname string, port int32) *ExecutorLib {
	fID := frameworkId
	eID := executorId
	hn := hotname
	p := port
	return &ExecutorLib{
		name:        name,
		agent:       &mesosproto.AgentInfo{Hostname: &hn, Port: &p},
		frameworkID: &mesosproto.FrameworkID{Value: &fID},
		executorID:  &mesosproto.ExecutorID{Value: &eID},
		tasks:       make(map[string]*mesosproto.TaskInfo),
	}
}

type TaskHandler func(task *mesosproto.TaskInfo, event executorproto.Event_Type)
