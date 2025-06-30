package main

import (
	"github.com/folivorra/goRedis/internal/app"
	"github.com/folivorra/goRedis/internal/config"
	"log"
	"os"
)

func main() {
	cfgPath := os.Getenv("APP_CONFIG")
	if cfgPath == "" {
		cfgPath = "/app/config/app_config.yaml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	a, err := app.NewApp(cfg)
	if err != nil {
		log.Fatal(err)
	}

	a.Start()
	a.Wait()
	a.Shutdown()

	// TODO: event-sourcing arch || state machine || SAGA pattern
}
