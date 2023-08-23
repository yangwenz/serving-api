package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yangwenz/model-serving/api"
	"github.com/yangwenz/model-serving/platform"
	"github.com/yangwenz/model-serving/utils"
	"os"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}
	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	// Initialize ML platform service
	service := platform.NewKServe(config)
	// Start model API server
	runGinServer(config, service)
}

func runGinServer(config utils.Config, platform platform.Platform) {
	server, err := api.NewServer(config, platform)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}
