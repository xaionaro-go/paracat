package main

import (
	"flag"
	"log"

	"github.com/chenx-dust/paracat/app"
	"github.com/chenx-dust/paracat/app/client"
	"github.com/chenx-dust/paracat/app/relay"
	"github.com/chenx-dust/paracat/app/server"
	"github.com/chenx-dust/paracat/config"
)

func main() {
	cfgFilename := flag.String("c", "config.json", "config file")
	flag.Parse()

	log.Println("loading config from", *cfgFilename)
	cfg, err := config.LoadFromFile(*cfgFilename)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	// log.Println("config loaded:", cfg)

	var application app.App
	if cfg.Mode == config.ClientMode {
		application = client.NewClient(cfg)
	} else if cfg.Mode == config.ServerMode {
		application = server.NewServer(cfg)
	} else if cfg.Mode == config.RelayMode {
		application = relay.NewRelay(cfg)
	} else {
		log.Fatalf("Invalid mode: %v", cfg.Mode)
	}

	err = application.Run()
	if err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
