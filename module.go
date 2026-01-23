package esp32wifi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	pb "go.viam.com/api/component/board/v1"
	board "go.viam.com/rdk/components/board"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var (
	Esp32Wifi = resource.NewModel("mattmacf", "esp32-wifi", "esp32-wifi")
)

func init() {
	resource.RegisterComponent(board.API, Esp32Wifi,
		resource.Registration[board.Board, *Config]{
			Constructor: newEsp32WifiEsp32Wifi,
		},
	)
}

type Config struct {
	Url string `json:"url"`
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
func (cfg *Config) Validate(path string) ([]string, []string, error) {
	if cfg.Url == "" {
		return nil, nil, fmt.Errorf("%s: missing required field 'url'", path)
	}
	return nil, nil, nil
}

type esp32WifiEsp32Wifi struct {
	resource.AlwaysRebuild

	name resource.Name

	logger logging.Logger
	cfg    *Config
	url    string

	cancelCtx  context.Context
	cancelFunc func()
}

func newEsp32WifiEsp32Wifi(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (board.Board, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewEsp32Wifi(ctx, deps, rawConf.ResourceName(), conf, logger)

}

func NewEsp32Wifi(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (board.Board, error) {

	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	s := &esp32WifiEsp32Wifi{
		name:       name,
		logger:     logger,
		cfg:        conf,
		url:        conf.Url,
		cancelCtx:  cancelCtx,
		cancelFunc: cancelFunc,
	}
	return s, nil
}

func (s *esp32WifiEsp32Wifi) Name() resource.Name {
	return s.name
}

// AnalogByName returns an analog pin by name.
func (s *esp32WifiEsp32Wifi) AnalogByName(name string) (board.Analog, error) {
	var analogRetVal board.Analog
	analogRetVal = &analogClient{
		esp32WifiEsp32Wifi: s,
		boardName:          s.name.ShortName(),
		analogName:         name,
	}

	return analogRetVal, nil
}

// DigitalInterruptByName returns a digital interrupt by name.
func (s *esp32WifiEsp32Wifi) DigitalInterruptByName(name string) (board.DigitalInterrupt, error) {
	var digitalInterruptRetVal board.DigitalInterrupt

	return digitalInterruptRetVal, fmt.Errorf("DigitalInterruptByName not implemented")
}

// GPIOPinByName returns a GPIOPin by name.
func (s *esp32WifiEsp32Wifi) GPIOPinByName(name string) (board.GPIOPin, error) {
	var gPIOPinRetVal board.GPIOPin
	gPIOPinRetVal = &gpioPinClient{
		esp32WifiEsp32Wifi: s,
		boardName:          s.name.ShortName(),
		pinName:            name,
	}

	return gPIOPinRetVal, nil
}

// SetPowerMode sets the board to the given power mode. If
// provided, the board will exit the given power mode after
// the specified duration.
func (s *esp32WifiEsp32Wifi) SetPowerMode(ctx context.Context, mode pb.PowerMode, duration *time.Duration, extra map[string]interface{}) error {
	return fmt.Errorf("SetPowerMode not implemented")
}

func (s *esp32WifiEsp32Wifi) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return nil, fmt.Errorf("DoCommand not implemented")
}

type analogClient struct {
	*esp32WifiEsp32Wifi
	boardName  string
	analogName string
}

func (s *analogClient) Read(ctx context.Context, extra map[string]interface{}) (board.AnalogValue, error) {
	var analogValueRetVal board.AnalogValue
	endpoint := fmt.Sprintf("%s/read-pins", s.url)
	pinNum, err := strconv.Atoi(s.analogName)
	if err != nil {
		return analogValueRetVal, fmt.Errorf("failed to convert pin name to number: %w", err)
	}
	body := map[string]interface{}{
		"pin_reads": []int{pinNum},
	}

	s.logger.Infof("using url: %s", endpoint)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return analogValueRetVal, fmt.Errorf("failed to marshal body: %w", err)
	}
	s.logger.Infof("jsonBody: %s", string(jsonBody))

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return analogValueRetVal, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return analogValueRetVal, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return analogValueRetVal, fmt.Errorf("failed to read pin: %s", resp.Status)
	}

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return analogValueRetVal, fmt.Errorf("failed to decode response: %w", err)
	}
	s.logger.Infof("response: %+v", response)

	state := response["pin_reads"].([]interface{})[0].(map[string]interface{})["state"].(float64)

	return board.AnalogValue{
		Value: int(state),
	}, nil
}

func (s *analogClient) Write(ctx context.Context, value int, extra map[string]interface{}) error {
	return fmt.Errorf("Write not implemented")
}

type digitalInterruptClient struct {
	*esp32WifiEsp32Wifi
	boardName            string
	digitalInterruptName string
}

func (s *digitalInterruptClient) Value(ctx context.Context, extra map[string]interface{}) (int64, error) {
	return 0, fmt.Errorf("Value not implemented")
}

// StreamTicks starts a stream of digital interrupt ticks.
func (s *esp32WifiEsp32Wifi) StreamTicks(ctx context.Context, interrupts []board.DigitalInterrupt, ch chan board.Tick, extra map[string]interface{}) error {
	return fmt.Errorf("StreamTicks not implemented")
}

type gpioPinClient struct {
	*esp32WifiEsp32Wifi
	boardName string
	pinName   string
}

func (s *gpioPinClient) Set(ctx context.Context, high bool, extra map[string]interface{}) error {
	state := 0
	if high {
		state = 100
	}
	endpoint := fmt.Sprintf("%s/write-pins", s.url)
	pinNum, err := strconv.Atoi(s.pinName)
	if err != nil {
		return fmt.Errorf("failed to convert pin name to number: %w", err)
	}
	body := map[string]interface{}{
		"pin_writes": []map[string]interface{}{
			{
				"pin_num": pinNum,
				"state":   state,
			},
		},
	}

	s.logger.Infof("using url: %s", endpoint)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}
	s.logger.Infof("jsonBody: %s", string(jsonBody))

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to write pin: %s", resp.Status)
	}

	return nil
}

func (s *gpioPinClient) Get(ctx context.Context, extra map[string]interface{}) (bool, error) {
	endpoint := fmt.Sprintf("%s/read-pins", s.url)
	pinNum, err := strconv.Atoi(s.pinName)
	if err != nil {
		return false, fmt.Errorf("failed to convert pin name to number: %w", err)
	}
	body := map[string]interface{}{
		"pin_reads": []int{pinNum},
	}

	s.logger.Infof("using url: %s", endpoint)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return false, fmt.Errorf("failed to marshal body: %w", err)
	}
	s.logger.Infof("jsonBody: %s", string(jsonBody))

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to read pin: %s", resp.Status)
	}

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}
	s.logger.Infof("response: %+v", response)

	return response["pin_reads"].([]interface{})[0].(map[string]interface{})["state"].(float64) == 100, nil
}

func (s *gpioPinClient) PWM(ctx context.Context, extra map[string]interface{}) (float64, error) {
	endpoint := fmt.Sprintf("%s/read-pins", s.url)
	pinNum, err := strconv.Atoi(s.pinName)
	if err != nil {
		return 0, fmt.Errorf("failed to convert pin name to number: %w", err)
	}
	body := map[string]interface{}{
		"pin_reads": []int{pinNum},
	}

	s.logger.Infof("using url: %s", endpoint)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal body: %w", err)
	}
	s.logger.Infof("jsonBody: %s", string(jsonBody))

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to read pin: %s", resp.Status)
	}

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}
	s.logger.Infof("response: %+v", response)

	return response["pin_reads"].([]interface{})[0].(map[string]interface{})["state"].(float64), nil
}

func (s *gpioPinClient) SetPWM(ctx context.Context, dutyCyclePct float64, extra map[string]interface{}) error {
	endpoint := fmt.Sprintf("%s/write-pins", s.url)
	pinNum, err := strconv.Atoi(s.pinName)
	if err != nil {
		return fmt.Errorf("failed to convert pin name to number: %w", err)
	}
	body := map[string]interface{}{
		"pin_writes": []map[string]interface{}{
			{
				"pin_num": pinNum,
				"state":   int(dutyCyclePct * 100),
			},
		},
	}

	s.logger.Infof("using url: %s", endpoint)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}
	s.logger.Infof("jsonBody: %s", string(jsonBody))

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to write pin: %s", resp.Status)
	}

	return nil
}

func (s *gpioPinClient) PWMFreq(ctx context.Context, extra map[string]interface{}) (uint, error) {
	return 0, fmt.Errorf("PWMFreq not implemented")
}

func (s *gpioPinClient) SetPWMFreq(ctx context.Context, freqHz uint, extra map[string]interface{}) error {
	return fmt.Errorf("SetPWMFreq not implemented")
}

func (s *esp32WifiEsp32Wifi) Close(context.Context) error {
	// Put close code here
	s.cancelFunc()
	return nil
}
