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

func (c *client) createOCIbundleAndRun(taskId, containerImage, args string, taskInfo *mesosproto.TaskInfo, restore bool) error {
	// create the top most bundle and rootfs directory
	log.Infof("Creating OCI bundle for %s with image: %s", taskId, containerImage)

	rootPath := filepath.Join(taskId, "rootfs")
	if err := os.MkdirAll(rootPath, 0777); err != nil {
		log.Infof("ERROR mkdir %#v:", err)
		log.Error(err)
		return err
	}

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
	log.Infof("Creating spec for: %#v", taskId)
	/*
		cmd = exec.Command("runc", "spec", "-b", taskId)
		err = cmd.Run()
		if err != nil {
			log.Infof("ERROR cmd exec %#v:", err)

			log.Error(err)
			return err
		}

		configPath := filepath.Join(taskId, "config.json")
		// Editing the spec sed  's/\"terminal\": true,/\"terminal\": false/' config.json
		log.Infof("Editing spec for: %#v", taskId)
		cmd = exec.Command("sh", "-c", fmt.Sprintf("sed -i 's/\"terminal\": true,/\"terminal\": false,/' %s", configPath))
		err = cmd.Run()
		if err != nil {
			log.Infof("ERROR cmd exec %#v:", err)

			log.Error(err)
			return err
		}

		// Editing the spec for command
		log.Infof("Editing spec for command: %s", args)
		cmd = exec.Command("sh", "-c", fmt.Sprintf("sed -i 's;\"sh\";\"%s\";' %s", args, configPath))
		err = cmd.Run()
		if err != nil {
			log.Infof("ERROR cmd exec %#v:", err)

			log.Error(err)
			return err
		}
	*/
	configFile, err := os.Create(filepath.Join(taskId, "config.json"))
	if err != nil {
		log.Infof("ERROR create file %#v:", err)
		log.Error(err)
		return err
	}
	_, err = fmt.Fprintf(configFile, OCI_CONFIG_TEMPLATE, args)
	configFile.Close()
	if err != nil {
		log.Infof("ERROR create file %#v:", err)
		log.Error(err)
		return err
	}

	// run container in runc
	log.Infof("Running container from image: %#v  with runc in: %#v", containerImage, taskId)

	go func() {
		if !restore {
			cmd = exec.Command("runc", "run", "-b", taskId, taskId)
		} else {
			cmd = exec.Command("runc", "restore", "--image-path", filepath.Join("/criu", taskId), "-b", taskId, taskId)
		}
		var stderr bytes.Buffer

		cmd.Stderr = &stderr

		err = cmd.Run()
		state := mesosproto.TaskState_TASK_FINISHED
		if err != nil {
			log.Infof("ERROR running : %s", fmt.Sprint(err)+": "+stderr.String())
			state = mesosproto.TaskState_TASK_FAILED
			// TODO: send status update task failed
		}
		if err := c.lib.Update(taskInfo, &state); err != nil {
			log.Errorf("Update task state as %s failed: %v", state.String(), err)
		}

	}()

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

func checkpointTask(data []byte) error {
	taskId := strings.Split(fmt.Sprintf("%s", data), " ")[1]
	cmd := exec.Command("runc", "checkpoint", "--image-path", "/criu/"+taskId, taskId)
	var stderr bytes.Buffer

	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		log.Infof("ERROR checkpointing : %s", fmt.Sprint(err)+": "+stderr.String())
		log.Error(err)
		return err
	}
	return nil
}

func (c *client) handleTasks(task *mesosproto.TaskInfo, event *executorproto.Event) error {
	c.Lock()
	defer c.Unlock()
	switch event.GetType() {
	case executorproto.Event_MESSAGE:
		data := event.GetMessage().GetData()
		log.Infof("MESSAGE RECEIVED with: %q", data)
		if err := checkpointTask(data); err != nil {
			log.Errorf("Executor checkpoint error: %v", err)
			return err
		}
	case executorproto.Event_LAUNCH:
		task := event.GetLaunch().GetTask()
		taskId := task.GetTaskId().GetValue()
		labels := task.GetLabels().GetLabels()
		restoring := false
		for _, label := range labels {
			if label.GetKey() == "restore" && label.GetValue() == taskId {
				restoring = true
			}
		}
		if containerType := task.GetContainer().GetType(); containerType == mesosproto.ContainerInfo_DOCKER {
			containerImage := task.GetContainer().GetDocker().GetImage()
			log.Infof("LAUNCH RECEIVED for task: %#v for image %#v", taskId, containerImage)
			args := task.GetContainer().GetDocker().GetParameters()[0].GetValue()

			c.createOCIbundleAndRun(taskId, containerImage, args, task, restoring)
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
	return nil
}

func main() {

	agent := flag.String("-agent", os.Getenv("AGENT_ADDR"), "Mesos Agent to connect to")

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
