package lib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gogo/protobuf/proto"
	"github.com/jimenez/mesoscon-demo/lib/mesosproto"
)

const ENDPOINT = "/master/api/v1/scheduler"

type DemoLib struct {
	name          string
	master        string
	frameworkInfo *mesosproto.FrameworkInfo
	frameworkID   *mesosproto.FrameworkID
	tasks         map[string]*mesosproto.AgentID
	OffersCH      chan *mesosproto.Offer
}

func New(master, name string) *DemoLib {
	return &DemoLib{
		name:          name,
		master:        master,
		frameworkInfo: &mesosproto.FrameworkInfo{Name: &name, User: proto.String("root")},
		tasks:         make(map[string]*mesosproto.AgentID),
		OffersCH:      make(chan *mesosproto.Offer),
	}
}

func (lib *DemoLib) handleEvents(body io.Reader) {
	dec := json.NewDecoder(body)
	for {
		var event mesosproto.Event

		if err := dec.Decode(&event); err != nil {
			fmt.Printf("%#v\n", event)
			if err == io.EOF {
				break
			}
			if event.GetType() == mesosproto.Event_UPDATE {
				//				fmt.Printf("%#v\n", event)
				taskStatus := event.GetUpdate().GetStatus()
				fmt.Printf("%#v\n", taskStatus)
				lib.tasks[taskStatus.GetTaskId().GetValue()] = taskStatus.GetAgentId()
				log.Println("Status for", taskStatus.GetTaskId().GetValue(), "on", taskStatus.GetAgentId().GetValue(), "is", taskStatus.GetState().String())
				if taskStatus.GetUuid() != nil {
					log.Println("Acknowledge", taskStatus.GetTaskId().GetValue())
					lib.Acknowledge(taskStatus.GetTaskId(), taskStatus.GetAgentId(), taskStatus.GetUuid())
				}
			}
			continue
		}

		switch event.GetType() {
		case mesosproto.Event_SUBSCRIBED:
			lib.frameworkID = event.GetSubscribed().GetFrameworkId()
			log.Println("framework", lib.name, "subscribed succesfully (", lib.frameworkID.String(), ")")
		case mesosproto.Event_OFFERS:
			for _, offer := range event.GetOffers().GetOffers() {
				lib.OffersCH <- offer
			}
			log.Println("framework", lib.name, "received", len(event.GetOffers().GetOffers()), "offer(s)")
		}
	}
}

func (lib *DemoLib) Subscribe() error {
	call := mesosproto.Call{
		Type: mesosproto.Call_SUBSCRIBE.Enum(),
		Subscribe: &mesosproto.Call_Subscribe{
			FrameworkInfo: lib.frameworkInfo,
		},
	}

	body, err := proto.Marshal(&call)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "http://"+lib.master+ENDPOINT, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("%s", body)
	}

	go lib.handleEvents(resp.Body)
	return nil

}
