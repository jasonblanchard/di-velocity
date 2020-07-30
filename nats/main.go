package main

import (
	"flag"
	"fmt"

	"github.com/jasonblanchard/di-velocity/container"
	"github.com/spf13/viper"

	_ "github.com/lib/pq"

	"github.com/jasonblanchard/natsby"
)

func main() {
	config := flag.String("config", "", "Path to config file")
	flag.Parse()

	configFile := container.InitExternalConfig(*config)

	containerInput := &container.Input{
		PostgresUser:     viper.GetString("db_user"),
		PostgresPassword: viper.GetString("db_password"),
		PostgresHost:     viper.GetString("db_host"),
		PostgresDbName:   viper.GetString("db_name"),
		NatsURL:          viper.GetString("nats_url"),
		Debug:            viper.GetBool("debug"),
		Pretty:           viper.GetBool("pretty"),
		TestMode:         viper.GetBool("test_mode"),
		NATSQueue:        "velocity",
	}

	container, err := container.New(containerInput)

	if err != nil {
		panic(err)
	}

	if configFile != "" {
		container.Logger.Info().Msg(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))
	}

	engine, err := natsby.New(container.NATSConnection)
	if err != nil {
		panic(err)
	}

	engine.Use(natsby.WithLogger(container.Logger))
	// engine.Use(natsby.WithCustomRecovery(Recovery(container)))

	SubscribeHandlers(container, engine)

	engine.Run(func() {
		container.Logger.Info().Msg("Ready to receive messages")
	})
}
