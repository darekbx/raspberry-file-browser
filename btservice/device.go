package cosmose

import (
	"fmt"
	"log"
	"time"
	"math"
	"strings"
	"os/exec"

	"github.com/paypal/gatt"
)

var serviceUuid = "09fc95c0-c111-11e3-9904-0002a5d5c51b"
var writeCharacteristicUuid = "16fe0d80-c111-11e3-b8c8-0002a5d5c51b"
var notifyCharacteristicUuid = "1c927b50-c116-11e3-8a33-0800200c9a66"
var scanJob = "/home/pi/projects/pi-scanner/job.sh"
var combinedJobParameter = " --combined"
var regionSetJob = "sh /home/pi/region_set.sh "
var restartJob = "/home/pi/restart.sh"
var chunkSize = 10.0
var scanWifi = false

func Reset() {
	scanWifi = false
	log.Println("Performing reset")
}

func printPackets(input string, chunkSize float64, n gatt.Notifier) {
	length := len(input)
	chunks := math.Ceil(float64(length) / chunkSize)
	for i := 0; i < int(chunks); i ++ {
		chunkSizeInt := int(chunkSize)
		start := i * chunkSizeInt
		end := int(math.Min(float64(length), float64(start) + chunkSize))
		chunk := input[start:end]
		fmt.Fprintf(n, "%s", chunk)
	}
}

func NewCosmoseService() *gatt.Service {
	noopDelay := 250 * time.Millisecond
	scanWifi := false
	isCombinedScan := false

	s := gatt.NewService(gatt.MustParseUUID(serviceUuid))
	s.AddCharacteristic(gatt.MustParseUUID(writeCharacteristicUuid)).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			log.Println("Action received: ", string(data))
			chunks := strings.Split(string(data), ":")
			action := chunks[0]
			var argument string
			if len(chunks) == 2 {
				argument = chunks[1]
				log.Println("Action argument: ", string(argument))
			}
			switch string(action) {
				case "start":
					scanWifi = true
					if argument != "" {
						isCombinedScan = true
					}
				case "stop":
					scanWifi = false
				case "region":
					log.Println("Setting wireless region to:", string(argument))
					regionErr := exec.Command("sh", "-c", regionSetJob + string(argument)).Run()
					if regionErr != nil {
						log.Println(regionErr)
					}
				case "restart":
					log.Println("Restarting server...")
					exec.Command("/bin/bash", restartJob).Run()
					log.Println("Restart completed")
			}
			return gatt.StatusSuccess
		})

	s.AddCharacteristic(gatt.MustParseUUID(notifyCharacteristicUuid)).HandleNotifyFunc(
		func(r gatt.Request, n gatt.Notifier) {
			for !n.Done() {
				if scanWifi {
					parameter := ""
					if isCombinedScan {
						log.Println("Combined scan in progress...")
						parameter = combinedJobParameter
						isCombinedScan = false
					} else {
						log.Println("scan in progress...")
					}
					out, err := exec.Command("/bin/bash", scanJob, parameter).Output()
					if err != nil {
						log.Println(err)
					} else {
						outString := string(out)
						fmt.Fprintf(n, "---start-%d", len(outString))
						printPackets(outString, chunkSize, n)
						fmt.Fprintf(n, "---end-%d", len(outString))
						log.Println("Printing packets...")
					}
				} else {
					log.Println("NOOP")
					time.Sleep(noopDelay)
				}
			}
		})

	return s
}
