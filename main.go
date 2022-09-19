package main

import (
	"flag"
	"log"

	"github.com/yanando/arcy_monitor/monitor"
)

func main() {
	var link = flag.String("l", "", "arc'teryx website link")
	var webhook = flag.String("w", "", "webhook")
	flag.Parse()

	if *link == "" {
		log.Fatal("Please provide a link to monitor (--help)")
	} else if *webhook == "" {
		log.Fatal("Please provide a valid discord webhook (--help)")
	}

	m := monitor.Monitor{
		Link:    *link,
		Webhook: *webhook,
	}

	m.Monitor()
}
