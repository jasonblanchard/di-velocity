package main

import (
	"flag"
	"fmt"

	"github.com/jasonblanchard/di-velocity/internal/container"
	"github.com/jasonblanchard/di-velocity/internal/handlers"
	"github.com/spf13/viper"

	_ "github.com/lib/pq"

	"github.com/jasonblanchard/natsby"
	nats "github.com/nats-io/nats.go"
)

var natsQueue = "valocity"

func initConfig(path string) string {
	if path != "" {
		viper.SetConfigFile(path)
	}
	viper.AutomaticEnv()
	err := viper.ReadInConfig()

	if err != nil {
		return ""
	}
	return viper.ConfigFileUsed()
}

func main() {
	config := flag.String("config", "", "Path to config file")
	flag.Parse()

	configFile := initConfig(*config)

	containerInput := &container.Input{
		PostgresUser:     viper.GetString("db_user"),
		PostgresPassword: viper.GetString("db_password"),
		PostgresDbName:   viper.GetString("db_name"),
		NatsURL:          nats.DefaultURL,
		Debug:            viper.GetBool("debug"),
		Pretty:           viper.GetBool("pretty"),
		TestMode:         viper.GetBool("test_mode"),
	}

	container, err := container.New(containerInput)

	if err != nil {
		panic(err)
	}

	if configFile != "" {
		container.Logger.Info().Msg(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))
	}

	configureLogger := func(e *natsby.Engine) error {
		e.Logger = container.Logger
		return nil
	}

	configureNATS := func(e *natsby.Engine) error {
		e.NatsConnection = container.Broker
		return nil
	}

	engine, err := natsby.New(configureLogger, configureNATS)
	if err != nil {
		panic(err)
	}

	engine.Use(natsby.WithLogger())

	handlers.Subscribe(container, engine)

	engine.Run(func() {
		container.Logger.Info().Msg("Ready to receive messages")
	})
}
