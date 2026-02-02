package esp32wifi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	pb "go.viam.com/api/component/board/v1"
	board "go.viam.com/rdk/components/board"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"tinygo.org/x/bluetooth"
)

var (
	Esp32Ble         = resource.NewModel("mattmacf", "esp32-wifi", "esp32-ble")
	errUnimplemented = errors.New("unimplemented")
	adapter          = bluetooth.DefaultAdapter
)

func init() {
	resource.RegisterComponent(board.API, Esp32Ble,
		resource.Registration[board.Board, *BleConfig]{
			Constructor: newEsp32BleEsp32Ble,
		},
	)
}

type BleConfig struct {
	BTServerName string `json:"bt_server_name"`
}

// Validate ensures all parts of the config are valid and important fields exist.
// Returns three values:
//  1. Required dependencies: other resources that must exist for this resource to work.
//  2. Optional dependencies: other resources that may exist but are not required.
//  3. An error if any Config fields are missing or invalid.
//
// The `path` parameter indicates
// where this resource appears in the machine's JSON configuration
// (for example, "components.0"). You can use it in error messages
// to indicate which resource has a problem.
func (cfg *BleConfig) Validate(path string) ([]string, []string, error) {
	if cfg.BTServerName == "" {
		return nil, nil, fmt.Errorf("%s: missing required field 'bt_server_name'", path)
	}
	return nil, nil, nil
}

type esp32BleEsp32Ble struct {
	resource.AlwaysRebuild

	name resource.Name

	logger       logging.Logger
	cfg          *BleConfig
	btServerName string
	device       *bluetooth.Device

	cancelCtx  context.Context
	cancelFunc func()
}

func newEsp32BleEsp32Ble(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (board.Board, error) {
	conf, err := resource.NativeConfig[*BleConfig](rawConf)
	if err != nil {
		return nil, err
	}

	return NewEsp32Ble(ctx, deps, rawConf.ResourceName(), conf, logger)

}

func NewEsp32Ble(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *BleConfig, logger logging.Logger) (board.Board, error) {

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	err := adapter.Enable()
	if err != nil {
		logger.Errorf("Failed to enable Bluetooth adapter: %v", err)
		cancelFunc()
		return nil, err
	}

	deviceFound := make(chan bluetooth.ScanResult, 1)
	timeout := time.After(10 * time.Second)

	// Start scanning
	go func() error {
		err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			deviceName := result.LocalName()

			// Print all discovered devices for visibility
			if deviceName != "" {
				logger.Infof("Found: %s (Address: %s, RSSI: %d dBm)",
					deviceName, result.Address.String(), result.RSSI)
			}

			// Check if this is the device we're looking for (case-insensitive)
			if strings.EqualFold(deviceName, conf.BTServerName) {
				select {
				case deviceFound <- result:
					adapter.StopScan()
				default:
				}
			}
		})

		if err != nil {
			logger.Errorf("Scan error: %v", err)
			cancelFunc()
			return err
		}
		return nil
	}()

	var device *bluetooth.Device

	// Wait for device to be found or timeout
	select {
	case result := <-deviceFound:
		logger.Infof("Found target device: %s", result.LocalName())
		logger.Infof("Address: %s", result.Address.String())
		logger.Infof("Signal strength: %d dBm", result.RSSI)

		// Connect to the device
		logger.Infof("Connecting...")

		dev, err := adapter.Connect(result.Address, bluetooth.ConnectionParams{})
		if err != nil {
			logger.Errorf("Failed to connect: %v", err)
			cancelFunc()
			return nil, err
		}
		device = &dev
	case <-timeout:
		logger.Errorf("Timeout waiting for device")
		cancelFunc()
		return nil, errors.New("timeout waiting for device")
	}

	s := &esp32BleEsp32Ble{
		name:         name,
		logger:       logger,
		cfg:          conf,
		btServerName: conf.BTServerName,
		device:       device,
		cancelCtx:    cancelCtx,
		cancelFunc:   cancelFunc,
	}
	//TODO: disconnect device when closing module
	return s, nil
}

func (s *esp32BleEsp32Ble) Name() resource.Name {
	return s.name
}

// AnalogByName returns an analog pin by name.
func (s *esp32BleEsp32Ble) AnalogByName(name string) (board.Analog, error) {
	var analogRetVal board.Analog

	return analogRetVal, fmt.Errorf("not implemented")
}

// DigitalInterruptByName returns a digital interrupt by name.
func (s *esp32BleEsp32Ble) DigitalInterruptByName(name string) (board.DigitalInterrupt, error) {
	var digitalInterruptRetVal board.DigitalInterrupt

	return digitalInterruptRetVal, fmt.Errorf("not implemented")
}

// GPIOPinByName returns a GPIOPin by name.
func (s *esp32BleEsp32Ble) GPIOPinByName(name string) (board.GPIOPin, error) {
	var gPIOPinRetVal board.GPIOPin
	gPIOPinRetVal = &bleGPIOPinClient{
		esp32BleEsp32Ble: s,
		boardName:        s.name.ShortName(),
		pinName:          name,
	}

	return gPIOPinRetVal, nil
}

// SetPowerMode sets the board to the given power mode. If
// provided, the board will exit the given power mode after
// the specified duration.
func (s *esp32BleEsp32Ble) SetPowerMode(ctx context.Context, mode pb.PowerMode, duration *time.Duration, extra map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}

func (s *esp32BleEsp32Ble) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

type bleAnalogClient struct {
	*esp32BleEsp32Ble
	boardName  string
	analogName string
}

func (s *bleAnalogClient) Read(ctx context.Context, extra map[string]interface{}) (board.AnalogValue, error) {
	var analogValueRetVal board.AnalogValue

	return analogValueRetVal, fmt.Errorf("not implemented")
}

func (s *bleAnalogClient) Write(ctx context.Context, value int, extra map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}

type bleDigitalInterruptClient struct {
	*esp32BleEsp32Ble
	boardName            string
	digitalInterruptName string
}

func (s *bleDigitalInterruptClient) Value(ctx context.Context, extra map[string]interface{}) (int64, error) {
	return 0, fmt.Errorf("not implemented")
}

// StreamTicks starts a stream of digital interrupt ticks.
func (s *esp32BleEsp32Ble) StreamTicks(ctx context.Context, interrupts []board.DigitalInterrupt, ch chan board.Tick, extra map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}

type bleGPIOPinClient struct {
	*esp32BleEsp32Ble
	boardName string
	pinName   string
}

func (s *bleGPIOPinClient) Set(ctx context.Context, high bool, extra map[string]interface{}) error {
	var targetChar bluetooth.DeviceCharacteristic
	var found bool

	// TODO: make so we dont neecd to do this every time
	services, err := s.device.DiscoverServices(nil)
	if err != nil {
		s.logger.Errorf("Failed to discover services: %v", err)
		return err
	}

	targetUUID, err := bluetooth.ParseUUID("c79b2ca7-f39d-4060-8168-816fa26737b7")
	if err != nil {
		s.logger.Errorf("Failed to parse UUID: %v", err)
		return err
	}
	for _, service := range services {
		chars, err := service.DiscoverCharacteristics([]bluetooth.UUID{targetUUID})
		if err != nil {
			continue
		}

		if len(chars) > 0 {
			targetChar = chars[0]
			found = true
			break
		}
	}
	if !found {
		s.logger.Errorf("Failed to find characteristic")
		return errors.New("failed to find characteristic")
	}

	state := 0
	if high {
		state = 100
	}
	pinNum, err := strconv.Atoi(s.pinName)
	body := map[string]interface{}{
		"pin_writes": []map[string]interface{}{
			{
				"pin_num": pinNum,
				"state":   state,
			},
		},
	}
	body_string, err := json.Marshal(body)

	targetChar.Write(body_string)
	return nil
}

func (s *bleGPIOPinClient) Get(ctx context.Context, extra map[string]interface{}) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (s *bleGPIOPinClient) PWM(ctx context.Context, extra map[string]interface{}) (float64, error) {
	return 0, fmt.Errorf("not implemented")
}

func (s *bleGPIOPinClient) SetPWM(ctx context.Context, dutyCyclePct float64, extra map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}

func (s *bleGPIOPinClient) PWMFreq(ctx context.Context, extra map[string]interface{}) (uint, error) {
	return 0, fmt.Errorf("not implemented")
}

func (s *bleGPIOPinClient) SetPWMFreq(ctx context.Context, freqHz uint, extra map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}

func (s *esp32BleEsp32Ble) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}
