package device

import (
	"fmt"
	"log"
	"math"
	"os/exec"
	"strings"
	"time"

	"github.com/paypal/gatt"
)

var serviceUUID = "00000000-1111-2222-3333-000000000001"
var writeCharacteristicUUID = "00000000-1111-2222-3333-000000000010"
var notifyCharacteristicUUID = "00000000-1111-2222-3333-000000000020"

var chunkStartHeader = "---start-"
var chunkEndHeader = "---end-"

var chunkSize = 20.0
var command = ""
var commandChunks = []string{}
var hasIncomingPacket = false

func printPackets(input string, chunkSize float64, n gatt.Notifier) {
	length := len(input)
	chunks := math.Ceil(float64(length) / chunkSize)
	for i := 0; i < int(chunks); i++ {
		chunkSizeInt := int(chunkSize)
		start := i * chunkSizeInt
		end := int(math.Min(float64(length), float64(start)+chunkSize))
		chunk := input[start:end]
		fmt.Fprintf(n, "%s", chunk)
	}
}

func mergePackets(packet string) {
	if strings.HasPrefix(packet, chunkStartHeader) {
		commandChunks = []string{}
		hasIncomingPacket = true
	} else if strings.HasPrefix(packet, chunkEndHeader) {
		hasIncomingPacket = false
		command = strings.Join(commandChunks, "")
	} else if hasIncomingPacket {
		commandChunks = append(commandChunks, packet)
	}
}

// NewService : create new bluetooth service
func NewService() *gatt.Service {
	noopDelay := 250 * time.Millisecond
	command := ""

	s := gatt.NewService(gatt.MustParseUUID(serviceUUID))
	s.AddCharacteristic(gatt.MustParseUUID(writeCharacteristicUUID)).HandleWriteFunc(
		func(r gatt.Request, data []byte) (status byte) {
			log.Println("Data received: ", string(data))
			mergePackets(string(data))
			return gatt.StatusSuccess
		})

	s.AddCharacteristic(gatt.MustParseUUID(notifyCharacteristicUUID)).HandleNotifyFunc(
		func(r gatt.Request, n gatt.Notifier) {
			for !n.Done() {

				if len(command) > 0 {
					out, err := exec.Command("/bin/bash", command).Output()
					command = ""
					outString := ""
					if err != nil {
						outString = string(err.Error())
					} else {
						outString = string(out)
					}

					fmt.Fprintf(n, "---start-%d", len(outString))
					printPackets(outString, chunkSize, n)
					fmt.Fprintf(n, "---end-%d", len(outString))
					log.Println("Printing packets...")
				}

				log.Println("NOOP")
				time.Sleep(noopDelay)
			}
		})

	return s
}
