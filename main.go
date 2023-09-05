package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yangwenz/model-serving/api"
	"github.com/yangwenz/model-serving/platform"
	"github.com/yangwenz/model-serving/utils"
	"github.com/yangwenz/model-serving/worker"
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
	webhook := platform.NewInternalWebhook(config)
	// Start task processor
	go runTaskProcessor(config, service)
	// Run task distributor
	distributor := worker.NewRedisTaskDistributor(config)
	// Start model API server
	runGinServer(config, service, distributor, webhook)
}

func runGinServer(
	config utils.Config,
	platform platform.Platform,
	distributor worker.TaskDistributor,
	webhook platform.Webhook,
) {
	server, err := api.NewServer(config, platform, distributor, webhook)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot start server")
	}
}

func runTaskProcessor(config utils.Config, platform platform.Platform) {
	if config.RedisAddress == "" {
		log.Fatal().Msg("redis address is not set")
	}
	taskProcessor := worker.NewRedisTaskProcessor(config, platform)
	log.Info().Msg("start task processor")
	err := taskProcessor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start task processor")
	}
}
