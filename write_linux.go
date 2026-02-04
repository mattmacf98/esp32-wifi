//go:build linux

package esp32wifi

import "tinygo.org/x/bluetooth"

func writeCharacteristic(char bluetooth.DeviceCharacteristic, data []byte) (int, error) {
	return char.WriteWithoutResponse(data)
}
