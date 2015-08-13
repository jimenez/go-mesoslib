package main

import (
	"bufio"
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
		line, hasMoreInLine, err := bio.ReadLine()
		if err != nil {
			log.Fatal(err)
		}
		if !hasMoreInLine {
			break
		}

		demoLib.LaunchTask(strings.TrimSpace(line))
	}
}
