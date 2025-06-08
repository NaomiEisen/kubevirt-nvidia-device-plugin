package main

import (
	"kubevirt-nvidia-device-plugin/pkg/device_plugin"
	"log"
)

func main() {
	log.Printf("Statring device plugin")
	device_plugin.InitiateDevicePlugin()
}
