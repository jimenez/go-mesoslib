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

func handleOffers(lib *lib.DemoLib) {
	for {
		offers = append(offers, <-lib.OffersCH)
	}
}

func main() {
	master := flag.String("-master", "localhost:5050", "Mesos Master to connect to")
	demoLib := lib.New(*master, "mesoscon-demo")
	if err := demoLib.Subscribe(); err != nil {
		log.Fatal(err)
	}
	go handleOffers(demoLib)

	stdin := bufio.NewReader(os.Stdin)
	for {
		line, _, err := stdin.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		id := make([]byte, 6)
		n, err := rand.Read(id)
		if n != len(id) || err != nil {
			continue
		}

		lline := bytes.NewBuffer(line).String()
		task := lib.Task{
			ID:      hex.EncodeToString(id),
			Command: strings.Split(strings.TrimSpace(lline), " "),
			Image:   "busybox",
		}
		offer := offers[0]

		offers = nil

		if err := demoLib.LaunchTask(offer, lib.BuildResources(0.1, 0, 0), &task); err != nil {
			log.Println(err)
		}
	}
}
