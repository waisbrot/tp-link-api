package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/waisbrot/tp-link-api/lib"
)

func main() {
	log.SetLevel(log.TraceLevel)
	conf := lib.Config{
		Host:     "http://192.168.0.1",
		Username: "admin",
		Password: "admin",
	}
	client, err := lib.NewClient(&conf)
	if err != nil {
		panic(err)
	}
	log.Info("Fetching address reservations")
	client.DHCPAddressReservations()
}
