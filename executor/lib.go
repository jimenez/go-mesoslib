package executor

import "github.com/jimenez/go-mesoslib/mesosproto"

const ENDPOINT = "/api/v1/executor"

type ExecutorLib struct {
	name             string
	agent            *mesosproto.AgentInfo
	frameworkID      *mesosproto.FrameworkID
	executorID       *mesosproto.ExecutorID
	tasks            map[string]*mesosproto.TaskInfo
	tasksUnAkowledge map[string]*mesosproto.TaskInfo
}

func New(agent, name, frameworkId, executorId string) *ExecutorLib {
	fID := frameworkId
	eID := executorId
	return &ExecutorLib{
		name:        name,
		frameworkID: &mesosproto.FrameworkID{Value: &fID},
		executorID:  &mesosproto.ExecutorID{Value: &eID},
		tasks:       make(map[string]*mesosproto.TaskInfo),
	}
}

type TaskHandler func(task *mesosproto.TaskInfo)
