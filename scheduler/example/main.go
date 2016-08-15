package main

import (
	"bufio"
	"bytes"
	"flag"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/jimenez/go-mesoslib"
	"github.com/jimenez/go-mesoslib/mesosproto"
	"github.com/jimenez/go-mesoslib/scheduler"
)

type client struct {
	sync.Mutex
	offers []*mesosproto.Offer
	lib    *scheduler.SchedulerLib
}

func (c *client) handleOffers(offer *mesosproto.Offer) {
	c.Lock()
	c.offers = append(c.offers, offer)
	c.Unlock()
}

func main() {
	master := flag.String("-master", "localhost:5050", "Mesos Master to connect to")
	demoClient := client{lib: scheduler.New(*master, "mesoscon-demo")}
	if err := demoClient.lib.Subscribe(demoClient.handleOffers); err != nil {
		log.Fatal(err)
	}

	stdin := bufio.NewReader(os.Stdin)
	for {
		line, _, err := stdin.ReadLine()
		if err != nil {
			log.Fatal(err)
		}

		array := strings.Split(strings.ToLower(strings.TrimSpace(bytes.NewBuffer(line).String())), " ")

		if len(array) < 1 || array[0] == "" {
			continue
		}

		switch array[0] {
		case "launch":
			if len(array) < 3 {
				log.Println("error: not enough parameters (launch <images> <cmd>)")
				continue
			}

			demoClient.Lock()
			if len(demoClient.offers) > 0 {
				offer := demoClient.offers[0]

				demoClient.offers = demoClient.offers[1:]

				if task := mesoslib.NewTask(array[1], array[2:]); task != nil {
					if err := demoClient.lib.LaunchTask(offer, mesoslib.BuildResources(0.1, 0, 0), task); err != nil {
						log.Println("error:", err)
					}
				}
			} else {
				log.Println("error: no offer available to start container")
			}
			demoClient.Unlock()
		case "kill":
			if len(array) < 2 {
				log.Println("error: not enough parameters (kill <taskId>)")
				continue
			}
			if err := demoClient.lib.KillTask(array[1]); err != nil {
				log.Println("error:", err)
			}
		default:
			log.Println("error: invalid command (launch, kill)")
		}
	}
}
