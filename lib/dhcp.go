package lib

import (
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
)

type DHCPAddressReservation struct {
	Id      uint64
	MAC     string
	IP      string
	Enabled bool
}

func (client *Client) DHCPAddressReservations() (reservations []DHCPAddressReservation, err error) {
	body, err := client.FetchPath("/userRpm/FixMapCfgRpm.htm")
	log.Debugf("%s", body)
	re := regexp.MustCompile(`"((?:[0-9A-F]{2}-){5}[0-9A-F]{2})", "([0-9.]+)", ([01]),`)
	matches := re.FindAllStringSubmatch(body, -1)
	log.Infof("Found %d address reservations", len(matches))
	for _, match := range matches {
		id, err := strconv.ParseUint(string(match[1][0]), 10, 64)
		if err != nil {
			return nil, err
		}
		mac := string(match[1][1])
		ip := string(match[1][2])
		enabled, err := strconv.ParseBool(string(match[1][3]))
		if err != nil {
			return nil, err
		}
		reservation := DHCPAddressReservation{
			Id:      id,
			MAC:     mac,
			IP:      ip,
			Enabled: enabled,
		}
		log.Debugf("reservation=%+v", reservation)
		reservations = append(reservations, reservation)
	}
	return
}
