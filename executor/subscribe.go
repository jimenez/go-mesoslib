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
			taskID := event.GetTaskId()
			taskInfo = lib.tasks[taskID.GetValue()]
			log.Println("Status for", taskID.GetValue(), "on agent: ", taskInfo.GetAgentId().GetValue(), "is acknowledged")
			delete(lib.tasksUnAkowledge, taskID.GetValue())
		}

		switch event.GetType() {
		case executorproto.Event_SUBSCRIBED:
			lib.agent = event.GetSubscribed().GetAgentInfo()
			log.Println("executor", lib.name, "subscribed succesfully (", lib.executorID.String(), ")")
		case executorproto.Event_MESSAGE:
			lib.message(event.GetData())
		case executorproto.Event_LAUNCH:
			taskInfo := event.GetTask()
			lib.tasks[tasInfo.GetTaskId().GetValue()] = taskInfo
			if err := lib.update(taskInfo.GetTaskId(), &mesosproto.TaskState_TASK_RUNNING); err != nil {
				logrus.Errorf("Update task state as RUNNING failed")
			}
			tasksUnAkowledge[taskInfo.GetTaskId().GetValue()] = taskInfo
			handler(taskInfo)
		case executorproto.Event_KILL:
			taskInfo := event.GetTask()
			delete(lib.tasks, tasInfo.GetTaskId().GetValue())
			if err := lib.update(taskInfo.GetTaskId(), &mesosproto.TaskState_TASK_KILLED); err != nil {
				logrus.Errorf("Update task state as KILLED failed")
			}
			tasksUnAkowledge[taskInfo.GetTaskId().GetValue()] = taskInfo
			handler(taskInfo)
		case executorproto.Event_ERROR:
			continue
		}
	}
}

func (lib *ExecutorLib) Subscribe(handler TaskHandler) error {
	call := &executorproto.Call{
		Type: executorproto.Call_SUBSCRIBE.Enum(),
		Subscribe: &executorproto.Call_Subscribe{
			FrameworkId: lib.frameworkID,
			ExecutorId:  lib.executorID,
		},
	}

	body, err := lib.send(call, 200)
	if err != nil {
		return err
	}
	go lib.handleEvents(body, handler)
	return nil
}
