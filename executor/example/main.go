package main

import (
	"flag"

	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/jimenez/go-mesoslib/executor"
	"github.com/jimenez/go-mesoslib/mesosproto"
	"github.com/jimenez/go-mesoslib/mesosproto/executorproto"
)

type client struct {
	sync.Mutex
	Tasks []*mesosproto.TaskInfo
	lib   *executor.ExecutorLib
}

func (c *client) handleTasks(task *mesosproto.TaskInfo, event *executorproto.Event) error {
	c.Lock()
	switch event.GetType() {
	case executorproto.Event_MESSAGE:
		log.Info("MESSAGE RECEIVED with: %v", event.GetMessage().GetData())
	case executorproto.Event_LAUNCH:
		log.Info("LAUNCH RECEIVED for task: %v", event.GetLaunch().GetTask().GetTaskId().GetValue())
	case executorproto.Event_KILL:
		log.Info("KILL RECEIVED for task: %v", event.GetKill().GetTaskId().GetValue())
	}
	c.Unlock()
	return nil
}

func main() {
	agent := flag.String("-agent", "localhost:5051", "Mesos Agent to connect to")

	frameworkID := flag.String("framework_id", "", "Id of Mesos Framework using the executor")
	executorID := flag.String("executor_id", "", "Id of Mesos Executor")

	flag.Parse()

	demoClient := client{lib: executor.New(*agent, "demo-executor", frameworkID, executorID)}

	if err := demoClient.lib.Subscribe(demoClient.handleTasks); err != nil {
		log.Fatal(err)
	}
}
