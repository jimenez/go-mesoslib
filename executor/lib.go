package executor

import (
	"github.com/jimenez/go-mesoslib/mesosproto"
	"github.com/jimenez/go-mesoslib/mesosproto/executorproto"
)

const ENDPOINT = "/api/v1/executor"

type ExecutorLib struct {
	name               string
	agent              string
	frameworkID        *mesosproto.FrameworkID
	executorID         *mesosproto.ExecutorID
	tasks              map[string]*mesosproto.TaskInfo
	tasksUnAcknowledge map[string]*mesosproto.TaskInfo
}

func New(agent, name string, frameworkId, executorId *string) *ExecutorLib {
	return &ExecutorLib{
		name:               name,
		agent:              agent,
		frameworkID:        &mesosproto.FrameworkID{Value: frameworkId},
		executorID:         &mesosproto.ExecutorID{Value: executorId},
		tasks:              make(map[string]*mesosproto.TaskInfo),
		tasksUnAcknowledge: make(map[string]*mesosproto.TaskInfo),
	}
}

type TaskHandler func(task *mesosproto.TaskInfo, event *executorproto.Event) error
