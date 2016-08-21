package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/jimenez/go-mesoslib/executor"
	"github.com/jimenez/go-mesoslib/mesosproto"
	"github.com/jimenez/go-mesoslib/mesosproto/executorproto"
)

type client struct {
	sync.Mutex
	tasks map[string]*mesosproto.TaskInfo
	lib   *executor.ExecutorLib
}

func createOCIbundleAndRun(taskId, containerImage string, args []string) error {
	// create the top most bundle and rootfs directory
	log.Infof("Creating OCI bundle for %s with image: %s", taskId, containerImage)

	// TODO: create path with something better than containerImage (ex: imageID or md5 of image)
	dirPath := filepath.Join(taskId, containerImage)
	rootPath := filepath.Join(dirPath, "rootfs")
	os.MkdirAll(rootPath, 0777)

	// export image via Docker into the rootfs directory
	log.Infof("Exporting image with Docker: %#v", containerImage)
	cmd := exec.Command("sh", "-c", fmt.Sprintf("docker export $(docker create %s)  | tar -C %s -xvf -", containerImage, rootPath))
	err := cmd.Run()
	if err != nil {
		log.Infof("ERROR cmd exec %#v:", err)
		log.Error(err)
		return err
	}

	// create runc spec
	log.Infof("Creating spec for: %#v", dirPath)
	cmd = exec.Command("runc", "spec", "-b", dirPath)
	err = cmd.Run()
	if err != nil {
		log.Infof("ERROR cmd exec %#v:", err)

		log.Error(err)
		return err
	}

	configPath := filepath.Join(dirPath, "config.json")
	// Editing the spec sed  's/\"terminal\": true,/\"terminal\": false/' config.json
	log.Infof("Editing spec for: %#v", dirPath)
	cmd = exec.Command("sh", "-c", fmt.Sprintf("sed -i 's/\"terminal\": true,/\"terminal\": false,/' %s", configPath))
	err = cmd.Run()
	if err != nil {
		log.Infof("ERROR cmd exec %#v:", err)

		log.Error(err)
		return err
	}

	// Editing the spec for command
	log.Infof("Editing spec for: %#v", dirPath)
	comnd := strings.Join(args, "\", \"")
	log.Infof("Command : %s", comnd)
	cmd = exec.Command("sh", "-c", fmt.Sprintf("sed -i 's;\"sh\";\"/%s\";' %s", comnd, configPath))
	err = cmd.Run()
	if err != nil {
		log.Infof("ERROR cmd exec %#v:", err)

		log.Error(err)
		return err
	}

	// run container in runc
	log.Infof("Running container from image: %#v  with runc in: %#v", containerImage, dirPath)

	// TODO: see comment for replacing the containerImage in the path
	// as it can collide.
	cmd = exec.Command("runc", "run", "-b", dirPath, containerImage)
	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil {
		log.Infof("ERROR running : $s", fmt.Sprint(err)+": "+stderr.String())
		log.Error(err)
		return err
	}
	return nil
}

func runcKill(taskId, containerImage string) error {
	// runc kill container

	// TODO: create path with something better than containerImage (ex: imageID or md5 of image)
	dirPath := filepath.Join(taskId, containerImage)

	cmd := exec.Command("runc", "kill", "-b", dirPath, containerImage)
	err := cmd.Run()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (c *client) handleTasks(task *mesosproto.TaskInfo, event *executorproto.Event) error {
	c.Lock()
	switch event.GetType() {
	case executorproto.Event_MESSAGE:
		log.Info("MESSAGE RECEIVED with: %v", event.GetMessage().GetData())
	case executorproto.Event_LAUNCH:
		task := event.GetLaunch().GetTask()
		taskId := task.GetTaskId().GetValue()
		if containerType := task.GetContainer().GetType(); containerType == mesosproto.ContainerInfo_DOCKER {
			containerImage := task.GetContainer().GetDocker().GetImage()
			log.Infof("LAUNCH RECEIVED for task: %#v for image %#v", taskId, containerImage)
			args := task.GetExecutor().GetCommand().GetArguments()
			createOCIbundleAndRun(taskId, containerImage, args)
		} else {
			log.Error("Executor only supports Docker containers")
		}
	case executorproto.Event_KILL:
		log.Info("KILL RECEIVED for task: %v", event.GetKill().GetTaskId().GetValue())
		taskId := task.GetTaskId().GetValue()
		if containerType := task.GetContainer().GetType(); containerType == mesosproto.ContainerInfo_DOCKER {
			containerImage := task.GetContainer().GetDocker().GetImage()
			runcKill(taskId, containerImage)
		} else {
			log.Error("Executor only supports Docker containers")
		}
	}
	c.Unlock()
	return nil
}

func main() {

	agent := flag.String("-agent", "localhost:5051", "Mesos Agent to connect to")

	frameworkID := flag.String("framework_id", "", "Id of Mesos Framework using the executor")
	executorID := flag.String("executor_id", "", "Id of Mesos Executor")

	flag.Parse()

	demoClient := client{
		lib:   executor.New(*agent, "demo-executor", frameworkID, executorID),
		tasks: make(map[string]*mesosproto.TaskInfo)}

	if err := demoClient.lib.Subscribe(demoClient.handleTasks); err != nil {
		log.Error(err)
	}

}
