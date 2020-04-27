package lib

import (
	"fmt"
	"github.com/tebeka/selenium"
	"strconv"
	"time"
)

type DHCPAddressReservation struct {
	Id      uint64
	MAC     string
	IP      string
	Enabled bool
}

func (client *Client) returnToMain() {
	// Focus back on the top frame should be enough to get to return to a neutral state
	client.driver.SwitchFrame(nil)
}

// Go to the DHCP reservations page. Assumes starting from the top frame. Ends up focused on the "main" frame (not the top frame!)
func (client *Client) dhcpReservationsPage() {
	client.driver.SwitchFrame("bottomLeftFrame")
	client.ClickElement("#a24")
	client.ClickElement("#a27")
	client.driver.SwitchFrame(nil)
	client.driver.SwitchFrame("mainFrame")
	time.Sleep(500)
}

const (
	cssDHCPTableMain           = "#autoWidth > tbody"
	cssDHCPTableAddresses      = cssDHCPTableMain + " > tr:nth-child(3) > td > table > tbody"
	cssDHCPAddNewButton        = cssDHCPTableMain + " > tr:nth-child(5) > td > input:nth-child(1)"
	cssDHCPAddNewMacAddrField  = cssDHCPTableMain + " > tr:nth-child(3) > td:nth-child(2) > input"
	cssDHCPAddNewIpAddrField   = cssDHCPTableMain + " > tr:nth-child(4) > td:nth-child(2) > input"
	cssDHCPAddNewEnabledSelect = cssDHCPTableMain + " > tr:nth-child(5) > td:nth-child(2) > select"
	cssDHCPAddNewSaveButton    = "input.buttonBig:nth-child(4)"
)

func cssDHCPTableAddressRow(id uint64) (selector string) {
	rowNum := id + 2
	return fmt.Sprintf("%s > tr:nth-child(%d)", cssDHCPTableAddresses, rowNum)
}

func dhcpReservationFromRow(row selenium.WebElement) (reservation DHCPAddressReservation, err error) {
	rowItems, err := row.FindElements(selenium.ByCSSSelector, "td")
	if err != nil {
		return
	}
	idTxt := getTextFromElement(rowItems[0])
	macTxt := getTextFromElement(rowItems[1])
	ipTxt := getTextFromElement(rowItems[2])
	enabledTxt := getTextFromElement(rowItems[3])
	idOut, err := strconv.ParseUint(idTxt, 10, 64)
	if err != nil {
		return
	}
	enabled := enabledTxt == "Enabled"
	reservation = DHCPAddressReservation{
		Id:      idOut,
		MAC:     macTxt,
		IP:      ipTxt,
		Enabled: enabled,
	}
	return
}

func (client *Client) DHCPAddressReservations() (reservations []DHCPAddressReservation, err error) {
	client.dhcpReservationsPage()
	defer client.returnToMain()
	rows := client.FindElementsBySelector(cssDHCPTableAddresses + "> tr")
	for _, row := range rows[1:] {
		reservation, err := dhcpReservationFromRow(row)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, reservation)
	}
	return
}

func (client *Client) CreateDHCPAddressReservation(reservation DHCPAddressReservation) (err error) {
	client.dhcpReservationsPage()
	defer client.returnToMain()
	client.ClickElement(cssDHCPAddNewButton)

	macField := client.FindElementBySelector(cssDHCPAddNewMacAddrField)
	sendKeysToElement(macField, reservation.MAC)

	ipField := client.FindElementBySelector(cssDHCPAddNewIpAddrField)
	sendKeysToElement(ipField, reservation.IP)

	enabledField := client.FindElementBySelector(cssDHCPAddNewEnabledSelect)
	var k string
	if reservation.Enabled {
		k = "e"
	} else {
		k = "d"
	}
	sendKeysToElement(enabledField, k)

	client.ClickElement(cssDHCPAddNewSaveButton)

	return nil
}
