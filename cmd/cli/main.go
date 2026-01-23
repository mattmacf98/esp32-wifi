package main

import (
	"context"
	"esp32wifi"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	board "go.viam.com/rdk/components/board"
)

func main() {
	err := realMain()
	if err != nil {
		panic(err)
	}
}

func realMain() error {
	ctx := context.Background()
	logger := logging.NewLogger("cli")

	deps := resource.Dependencies{}
	// can load these from a remote machine if you need

	cfg := esp32wifi.Config{}

	thing, err := esp32wifi.NewEsp32Wifi(ctx, deps, board.Named("foo"), &cfg, logger)
	if err != nil {
		return err
	}
	defer thing.Close(ctx)

	return nil
}
