package main

import (
	"flag"
	"log"
	"sync"

	"github.com/jimenez/go-mesoslib/executor"
	"github.com/jimenez/go-mesoslib/mesosproto"
	"github.com/jimenez/go-mesoslib/mesosproto/executorproto"
)

type client struct {
	sync.Mutex
	Tasks []*mesosproto.TaskInfo
	lib   *executor.ExecutorLib
}

func (c *client) handleTasks(task *mesosproto.TaskInfo, event executorproto.Event_Type) {
	c.Lock()
	switch event {
	case executorproto.Event_MESSAGE:

	case executorproto.Event_LAUNCH:
	case executorproto.Event_KILL:
	}
	c.Unlock()
}

func main() {
	agent := flag.String("-agent", "localhost:5051", "Mesos Agent to connect to")

	frameworkID := flag.String("-framework_id", "", "Id of Mesos Framework using the executor")
	executorID := flag.String("-executor_id", "", "Id of Mesos Executor")
	hostname := "localhost"
	port := "5051"

	demoClient := client{lib: executor.New("demo-executor", *frameworkID, *executorID, hostname, port)}
	if err := demoClient.lib.Subscribe(demoClient.handleTasks); err != nil {
		log.Fatal(err)
	}
}
