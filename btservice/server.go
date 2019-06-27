// +build

package main

import (
	"fmt"
	"log"

	"io/ioutil"
	"github.com/paypal/gatt"
	"github.com/paypal/gatt/linux/cmd"
	"cosmoseservice"
)

var DefaultServerOptions = []gatt.Option{
	gatt.LnxMaxConnections(1),
	gatt.LnxDeviceID(-1, true),
	gatt.LnxSetAdvertisingParameters(&cmd.LESetAdvertisingParameters{
		AdvertisingIntervalMin: 0x00f4,
		AdvertisingIntervalMax: 0x00f4,
		AdvertisingChannelMap:  0x7,
	}),
}

func main() {
	deviceName, err := ioutil.ReadFile("/home/pi/device_name")
    	if err != nil {
        	fmt.Print(err)
    	}
	d, err := gatt.NewDevice(DefaultServerOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s", err)
	}

	d.Handle(
		gatt.CentralConnected(func(c gatt.Central) { fmt.Printf("Connect to %s\n", string(deviceName)) }),
		gatt.CentralDisconnected(func(c gatt.Central) {
			fmt.Printf("Disconnect device %s\n ", string(deviceName))
			cosmose.Reset()
		}),
	)

	onStateChanged := func(d gatt.Device, s gatt.State) {
		fmt.Printf("State: %s\n", s)
		switch s {
		case gatt.StatePoweredOn:
			s1 := cosmose.NewCosmoseService()
			d.AddService(s1)
			d.AdvertiseNameAndServices(string(deviceName), []gatt.UUID{s1.UUID()})
		default:
		}
	}

	d.Init(onStateChanged)
	select {}
}
