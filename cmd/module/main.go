package main

import (
	"esp32wifi"

	board "go.viam.com/rdk/components/board"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
)

func main() {
	// ModularMain can take multiple APIModel arguments, if your module implements multiple models.
	module.ModularMain(resource.APIModel{board.API, esp32wifi.Esp32Wifi}, resource.APIModel{board.API, esp32wifi.Esp32Ble})
}
