package lib

import (
	"encoding/json"
	"io"
	"log"

	"github.com/jimenez/mesoscon-demo/lib/mesosproto"
	"github.com/samalba/dockerclient"
)

func (lib *DemoLib) handleEvents(body io.ReadCloser, handler OfferHandler) {
	dec := json.NewDecoder(body)
	for {
		var event mesosproto.Event
		if err := dec.Decode(&event); err != nil || event.Type == nil {
			continue
		}
		if event.GetType() == mesosproto.Event_UPDATE {
			taskStatus := event.GetUpdate().GetStatus()
			lib.tasks[taskStatus.GetTaskId().GetValue()] = taskStatus.GetAgentId()
			inspect := []dockerclient.ContainerInfo{}
			data := taskStatus.Data
			if data != nil && json.Unmarshal(data, &inspect) == nil && len(inspect) == 1 {

				lib.containers[taskStatus.GetTaskId().GetValue()] = inspect[0].Id
			}

			log.Println("Status for", taskStatus.GetTaskId().GetValue(), "on", taskStatus.GetAgentId().GetValue(), "is", taskStatus.GetState().String())
			if taskStatus.GetUuid() != nil {
				lib.Acknowledge(taskStatus.GetTaskId(), taskStatus.GetAgentId(), taskStatus.GetUuid())

			}
		}

		switch event.GetType() {
		case mesosproto.Event_SUBSCRIBED:
			lib.frameworkID = event.GetSubscribed().GetFrameworkId()
			log.Println("framework", lib.name, "subscribed succesfully (", lib.frameworkID.String(), ")")
		case mesosproto.Event_OFFERS:
			for _, offer := range event.GetOffers().GetOffers() {
				handler(offer)
			}
			log.Println("framework", lib.name, "received", len(event.GetOffers().GetOffers()), "offer(s)")
		}
	}
}

func (lib *DemoLib) Subscribe(handler OfferHandler) error {
	call := &mesosproto.Call{
		Type: mesosproto.Call_SUBSCRIBE.Enum(),
		Subscribe: &mesosproto.Call_Subscribe{
			FrameworkInfo: lib.frameworkInfo,
		},
	}

	body, err := lib.send(call, 200)
	if err != nil {
		return err
	}
	go lib.handleEvents(body, handler)
	return nil
}
