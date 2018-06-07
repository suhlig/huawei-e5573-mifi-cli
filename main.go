package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/tombuildsstuff/huawei-e5573-mifi-sdk-go/mifi"
)

var versionNumber = "0.1.0"

type MifiInformation struct {
	Carrier mifi.Carrier
	Status  mifi.Status
	Traffic mifi.TrafficStatistics
	Wifi    mifi.WifiSettings
}

func main() {
	endpoint := flag.String("endpoint", "http://192.168.1.1", "The endpoint of the Mifi. Defaults to `http://192.168.1.1`")
	showVersion := flag.Bool("version", false, "Display the Application Version")
	showHelp := flag.Bool("help", false, "Displays this message")

	flag.Parse()

	if *showHelp {
		flag.Usage()
		return
	}

	if *showVersion {
		log.Printf("v%s", versionNumber)
		return
	}

	m := mifi.Mifi{
		Endpoint: *endpoint,
	}

	err := run(m)
	if err != nil {
		panic(err)
	}
}

func run(m mifi.Mifi) error {
	info, err := populateMifiInformation(m)
	if err != nil {
		return fmt.Errorf("Error retrieving information for Mifi %q: %s", m.Endpoint, err)
	}

	output := info.format()
	println(output)

	return nil
}

func populateMifiInformation(m mifi.Mifi) (*MifiInformation, error) {
	err := m.ParseCookie()
	if err != nil {
		return nil, fmt.Errorf("Error obtaining authentication cookie for Mifi: %+v", err)
	}

	wifiSettings, err := m.WifiSettings()
	if err != nil {
		return nil, fmt.Errorf("Error getting Wifi Settings from the Mifi: %+v", err)
	}

	carrier, err := m.CarrierDetails()
	if err != nil {
		return nil, fmt.Errorf("Error getting Carrier Details from the Mifi: %+v", err)
	}

	status, err := m.CurrentStatus()
	if err != nil {
		return nil, fmt.Errorf("Error getting Status from the Mifi: %+v", err)
	}

	traffic, err := m.TrafficStatistics()
	if err != nil {
		return nil, fmt.Errorf("Error retrieving Traffic Statistics: %+v", err)
	}

	info := MifiInformation{
		Carrier: *carrier,
		Status:  *status,
		Traffic: *traffic,
		Wifi:    *wifiSettings,
	}
	return &info, nil
}

func (mi MifiInformation) format() string {
	network := buildNetworkInformation(mi.Carrier, mi.Status, mi.Traffic)
	info := buildGeneralInformation(mi.Status, mi.Wifi)
	return fmt.Sprintf(`
Mifi Status:
%s
%s
`, network, info)
}

func buildNetworkInformation(c mifi.Carrier, s mifi.Status, t mifi.TrafficStatistics) string {
	minutesConnected := t.SecondsConnectedToNetwork / 60
	hoursConnected := minutesConnected / 60
	str := `
  Network:
    Signal Strength: %d/%d bars
    Network:         %q (ID: %d)
    Bandwidth used:  %.2fMB down / %.2fMB up
    Connected for:   %d hours (%d minutes)`
	return fmt.Sprintf(str,
		s.CurrentSignalBars, s.MaxSignalBars,
		c.FullName, c.CarrierID,
		t.DownloadedMB, t.UploadedMB,
		hoursConnected, minutesConnected)
}

func buildGeneralInformation(s mifi.Status, w mifi.WifiSettings) string {
	return fmt.Sprintf(`
  Information:
    Battery: %d%%
    Wifi: %q (Country: %s | %d devices connected)
`, s.CurrentBatteryPercentage, w.SSID, w.Country, s.NumberOfUsersConnectedToWifi)
}