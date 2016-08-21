package executor

import (
	"encoding/json"
	"io"
	"log"

	"github.com/Sirupsen/logrus"
	"github.com/jimenez/go-mesoslib/mesosproto"
	"github.com/jimenez/go-mesoslib/mesosproto/executorproto"
)

func (lib *ExecutorLib) handleEvents(body io.ReadCloser, handler TaskHandler) {
	dec := json.NewDecoder(body)
	for {
		var event executorproto.Event
		if err := dec.Decode(&event); err != nil || event.Type == nil {
			continue
		}
		if event.GetType() == executorproto.Event_ACKNOWLEDGED {
			taskID := event.GetAcknowledged().GetTaskId()
			taskInfo := lib.tasks[taskID.GetValue()]
			log.Println("Status for", taskID.GetValue(), "on agent: ", taskInfo.GetAgentId().GetValue(), "is acknowledged")
			// TODO: mutex this instruction => delete(lib.tasksUnAkowledge, taskID.GetValue())
		}

		switch event.GetType() {
		case executorproto.Event_SUBSCRIBED:
			// TODO: check that event.GetSubscribed().GetAgentInfo() corresponds to correct info
			log.Println("executor", lib.name, "subscribed succesfully (", lib.executorID.String(), ")")
		case executorproto.Event_MESSAGE:
			if err := handler(nil, &event); err != nil {
				logrus.Error(err)
			}
		case executorproto.Event_LAUNCH:
			taskInfo := event.GetLaunch().GetTask()
			lib.tasks[taskInfo.GetTaskId().GetValue()] = taskInfo
			state := mesosproto.TaskState_TASK_RUNNING
			if err := handler(taskInfo, &event); err != nil {
				logrus.Error(err)
				state = mesosproto.TaskState_TASK_ERROR
			}
			if err := lib.Update(taskInfo, &state); err != nil {
				logrus.Errorf("Update task state as %s failed: %v", state.String(), err)
			}
			lib.tasksUnAcknowledge[taskInfo.GetTaskId().GetValue()] = taskInfo

		case executorproto.Event_KILL:
			taskID := event.GetKill().GetTaskId()
			taskInfo := lib.tasks[taskID.GetValue()]
			state := mesosproto.TaskState_TASK_KILLED
			if err := handler(taskInfo, &event); err != nil {
				logrus.Error(err)
				state = mesosproto.TaskState_TASK_ERROR

			}
			if err := lib.Update(taskInfo, &state); err != nil {
				logrus.Errorf("Update task state as %s failed: %v", state.String(), err)
			}
			lib.tasksUnAcknowledge[taskInfo.GetTaskId().GetValue()] = taskInfo

			// TODO: mutex this delete in a task.go method
			delete(lib.tasks, taskInfo.GetTaskId().GetValue())
		}
	}
}

func (lib *ExecutorLib) Subscribe(handler TaskHandler) error {
	call := &executorproto.Call{
		Type:        executorproto.Call_SUBSCRIBE.Enum(),
		Subscribe:   &executorproto.Call_Subscribe{},
		FrameworkId: lib.frameworkID,
		ExecutorId:  lib.executorID,
	}

	body, err := lib.send(call, 200)
	if err != nil {
		return err
	}
	lib.handleEvents(body, handler)
	return nil
}
