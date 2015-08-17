package lib

import (
	"log"

	"github.com/gogo/protobuf/proto"
	"github.com/jimenez/mesoscon-demo/lib/mesosproto"
	"github.com/jimenez/mesoscon-demo/lib/transport"
)

const ENDPOINT = "/master/api/v1/scheduler"

type DemoLib struct {
	name          string
	master        string
	frameworkInfo *mesosproto.FrameworkInfo
	frameworkID   *mesosproto.FrameworkID
	OffersCH      chan *mesosproto.Offer
}

func New(master, name string) *DemoLib {
	return &DemoLib{
		name:          name,
		master:        master,
		frameworkInfo: &mesosproto.FrameworkInfo{Name: &name, User: proto.String("root")},
		OffersCH:      make(chan *mesosproto.Offer),
	}
}

func (lib *DemoLib) handleEvents(s transport.Subscription) {
	ech := s.Events()
	for event := range ech {
		switch event.GetType() {
		case mesosproto.Event_UPDATE:
			taskStatus := event.GetUpdate().GetStatus()
			log.Println("Status for", taskStatus.GetTaskId().GetValue(), "is", taskStatus.GetState().String())
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
	log.Println("subscription terminated:", s.Err())
}

func (lib *DemoLib) Subscribe() error {
	s, err := transport.Subscribe(lib.master, lib.frameworkInfo, false)
	if err != nil {
		return err
	}
	go lib.handleEvents(s)
	return nil

}
