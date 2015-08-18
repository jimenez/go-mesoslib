package lib

import (
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
}

func New(master, name string) *DemoLib {
	return &DemoLib{
		name:          name,
		master:        master,
		frameworkInfo: &mesosproto.FrameworkInfo{Name: &name, User: proto.String("root")},
		tasks:         make(map[string]*mesosproto.AgentID),
	}
}

type OfferHandler func(offer *mesosproto.Offer)
