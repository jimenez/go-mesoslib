package lib

import "github.com/jimenez/mesoscon-demo/lib/mesosproto"

func (lib *DemoLib) LaunchTask(offer *mesosproto.Offer, resources []*mesosproto.Resource, task *Task) error {
	taskInfo := createTaskInfo(offer, resources, task)

	call := &mesosproto.Call{
		FrameworkId: lib.frameworkID,
		Type:        mesosproto.Call_ACCEPT.Enum(),
		Accept: &mesosproto.Call_Accept{
			OfferIds: []*mesosproto.OfferID{
				offer.Id,
			},
			Operations: []*mesosproto.Offer_Operation{
				&mesosproto.Offer_Operation{
					Type: mesosproto.Offer_Operation_LAUNCH.Enum(),
					Launch: &mesosproto.Offer_Operation_Launch{
						TaskInfos: []*mesosproto.TaskInfo{taskInfo},
					},
				},
			},
		},
	}

	_, err := lib.send(call, 202)
	return err
}
