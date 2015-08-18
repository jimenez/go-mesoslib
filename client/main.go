package main

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/jimenez/mesoscon-demo/lib"
	"github.com/jimenez/mesoscon-demo/lib/mesosproto"
)

var offers []*mesosproto.Offer

func handleOffers(offer *mesosproto.Offer) {
	offers = append(offers, offer)
}

func main() {
	master := flag.String("-master", "localhost:5050", "Mesos Master to connect to")
	demoLib := lib.New(*master, "mesoscon-demo")
	if err := demoLib.Subscribe(handleOffers); err != nil {
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
			id := make([]byte, 6)
			n, err := rand.Read(id)
			if n != len(id) || err != nil {
				continue
			}

			task := lib.Task{
				ID:      hex.EncodeToString(id),
				Command: array[2:],
				Image:   array[1],
			}
			if len(offers) == 0 {
				log.Println("error: no offer available to start container")
				continue
			}
			offer := offers[0]

			offers = offers[1:]

			if err := demoLib.LaunchTask(offer, lib.BuildResources(0.1, 0, 0), &task); err != nil {
				log.Println("error:", err)
			}
		case "kill":
			if len(array) < 2 {
				log.Println("error: not enough parameters (kill <taskId>)")
				continue
			}
			if err := demoLib.KillTask(array[1]); err != nil {
				log.Println("error:", err)
			}
		default:
			log.Println("error: invalid command (launch, kill)")
		}
	}
}
