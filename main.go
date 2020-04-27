package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/waisbrot/tp-link-api/lib"
)

func main() {
	log.SetLevel(log.InfoLevel)
	conf := lib.Config{
		Host:     "http://192.168.0.1",
		Username: "admin",
		Password: "admin",
	}
	client, err := lib.NewClient(&conf)
	defer client.Exit()
	if err != nil {
		panic(err)
	}
	log.Info("Fetching address reservations")
	rs, err := client.DHCPAddressReservations()
	if err != nil {
		panic(err)
	}
	log.Infof("Reservations: %+v", rs)
}
