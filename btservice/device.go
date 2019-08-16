package device

import (
	"bytes"
	"compress/gzip"
	b64 "encoding/base64"
	"fmt"
	"log"
	"math"
	"os/exec"
	"strings"
	"time"

	"github.com/paypal/gatt"
)

var serviceUUID = "eabea763-8144-4652-a831-82fc9d4e645c"
var writeCharacteristicUUID = "2e7a6c4b-b70e-49b6-acf9-2be297ac29e9"
var notifyCharacteristicUUID = "25db62b2-00d3-4df9-b7e0-125623d67008"

var chunkStartHeader = "---start-"
var chunkEndHeader = "---end-"

var chunkSize = 20.0
var commandChunks = []string{}
var hasIncomingPacket = false

// Command container
var Command string

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

func compress(input string) string {
	var buffer bytes.Buffer
	gz := gzip.NewWriter(&buffer)
	if _, err := gz.Write([]byte(input)); err != nil {
		return string(err.Error())
	}
	if err := gz.Flush(); err != nil {
		return string(err.Error())
	}
	if err := gz.Close(); err != nil {
		return string(err.Error())
	}
	return b64.StdEncoding.EncodeToString(buffer.Bytes())
}

func mergePackets(packet string) {
	if strings.HasPrefix(packet, chunkStartHeader) {
		commandChunks = []string{}
		hasIncomingPacket = true
	} else if strings.HasPrefix(packet, chunkEndHeader) {
		hasIncomingPacket = false
		Command = strings.Join(commandChunks, "")
		log.Println("Merged command: ", string(Command))
	} else if hasIncomingPacket {
		commandChunks = append(commandChunks, packet)
	}
}

// NewService : create new bluetooth service
func NewService() *gatt.Service {
	noopDelay := 250 * time.Millisecond
	Command = ""

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

				if len(Command) > 0 {
					log.Println("Execute: ", string(Command))
					out, err := exec.Command("/bin/sh", "-c", Command).Output()
					Command = ""
					outString := ""
					// Errors are raw string, valid output is compressed string in base64 format
					if err != nil {
						outString = string(err.Error())
					} else {
						outString = compress(string(out))
					}
					log.Println("Output: ", outString)
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
