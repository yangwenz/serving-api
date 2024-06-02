package main

import (
	"context"
	"github.com/HyperGAI/serving-api/api"
	"github.com/HyperGAI/serving-api/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	utils.InitZerolog()
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}
	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	webhook := api.NewInternalWebhook(config)
	runGinServer(config, webhook)
}

func runGinServer(
	config utils.Config,
	webhook api.Webhook,
) {
	server, err := api.NewServer(config, webhook)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create server")
	}
	/*
		err = server.Start(config.HTTPServerAddress)
		if err != nil {
			log.Fatal().Err(err).Msg("cannot start server")
		}
	*/
	httpServer := &http.Server{
		Addr:    config.HTTPServerAddress,
		Handler: server.Handler(),
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("cannot start server")
		}
	}()

	// https://gin-gonic.com/docs/examples/graceful-restart-or-stop/
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server shutdown")
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		log.Info().Msg("timeout of 5 seconds")
	}
	log.Info().Msg("server exiting")
}
