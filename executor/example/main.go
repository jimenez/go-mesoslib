package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

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

func createOCIbundleAndRun(taskId, containerImage string) error {
	// create the top most bundle and rootfs directory
	log.Infof("Creating OCI bundle for %s with image: %s", taskId, containerImage)
	dirPath := fmt.Sprintf("%s/%s", taskId, containerImage)
	os.MkdirAll(dirPath+"rootfs", 0777)

	// export image via Docker into the rootfs directory
	log.Infof("Exporting image with Docker: %#v", containerImage)
	cmd := exec.Command("sh", "-c", "docker export $(docker create "+containerImage+")  | tar -C "+dirPath+"/rootfs -xvf -")
	// cmd2 := exec.Command("tar", "-C", dirPath+"/rootfs", "-xvf", "-")

	// stdout, err := cmd.StdoutPipe()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// cmd2.Stdin = stdout
	// cmd2.Stdout = os.Stdout
	// _ = cmd2.Start()
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
		return err
	}
	//	_ = cmd2.Wait()

	// create runc spec
	log.Info("Exporting image with Docker: %#v", dirPath)
	cmd = exec.Command("runc", "spec", "-b", dirPath)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
		return err
	}

	// run container in runc
	log.Info("Running container from image: %#v  with runc in: %#v", containerImage, dirPath)
	cmd = exec.Command("runc", "run", "-b", dirPath, containerImage)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Info("ALL RUNNING")
	return nil
}

func runcKill(taskId, containerImage string) error {
	// runc kill container
	dirPath := fmt.Sprintf("/tmp/%s/%s", taskId, containerImage)
	cmd := exec.Command(fmt.Sprintf("runc kill -b %s %s", dirPath, containerImage))
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
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
			createOCIbundleAndRun(taskId, containerImage)
		} else {
			log.Fatal("Executor only supports Docker containers")
		}
	case executorproto.Event_KILL:
		log.Info("KILL RECEIVED for task: %v", event.GetKill().GetTaskId().GetValue())
		taskId := task.GetTaskId().GetValue()
		if containerType := task.GetContainer().GetType(); containerType == mesosproto.ContainerInfo_DOCKER {
			containerImage := task.GetContainer().GetDocker().GetImage()
			runcKill(taskId, containerImage)
		} else {
			log.Fatal("Executor only supports Docker containers")
		}
	}
	c.Unlock()
	return nil
}

func checkRunCinstalledorInstall() error {
	cmd := exec.Command("runc", "-h")
	err := cmd.Run()
	if err != nil {
		log.Info(" runc v1.0.0-rc1 DID NOT WORK")
		log.Fatal(err)
		return err
		log.Info("Fetching runc v1.0.0-rc1")
		cmd = exec.Command("wget", "https://github.com/opencontainers/runc/releases/download/v1.0.0-rc1/runc-linux-amd64")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
			return err
		}
		os.Chmod("runc-linux-amd64", 111)
		cmd = exec.Command("install", "-D", "-m0755", "runc-linux-amd64", "runc")
		err = cmd.Run()
		if err != nil {
			log.Fatal(err)
			return err
		}
	}
	return nil
}

func main() {

	if err := checkRunCinstalledorInstall(); err != nil {
		log.Fatal(err)
	}

	agent := flag.String("-agent", "localhost:5051", "Mesos Agent to connect to")

	frameworkID := flag.String("framework_id", "", "Id of Mesos Framework using the executor")
	executorID := flag.String("executor_id", "", "Id of Mesos Executor")

	flag.Parse()

	demoClient := client{
		lib:   executor.New(*agent, "demo-executor", frameworkID, executorID),
		tasks: make(map[string]*mesosproto.TaskInfo)}

	if err := demoClient.lib.Subscribe(demoClient.handleTasks); err != nil {
		log.Fatal(err)
	}

}
