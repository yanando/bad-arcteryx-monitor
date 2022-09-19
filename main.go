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
		Link:    "https://arcteryx.com/nl/en/shop/mens/beta-lt-jacket",
		Webhook: "https://discord.com/api/webhooks/1021479484370731098/qgFBZX5mbRB-s70Al0rjw2UfzbJo12dvRk0zMsjEzRW0NQ_tKuxUYxUSr1UIlSAF9dcG",
	}

	m.Monitor()
}
