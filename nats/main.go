package main

import (
	"flag"
	"fmt"

	"github.com/jasonblanchard/di-velocity/container"
	"github.com/spf13/viper"

	_ "github.com/lib/pq"

	"github.com/jasonblanchard/natsby"
	nats "github.com/nats-io/nats.go"
)

func main() {
	config := flag.String("config", "", "Path to config file")
	flag.Parse()

	configFile := container.InitExternalConfig(*config)

	containerInput := &container.Input{
		PostgresUser:     viper.GetString("db_user"),
		PostgresPassword: viper.GetString("db_password"),
		PostgresDbName:   viper.GetString("db_name"),
		NatsURL:          nats.DefaultURL,
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
	engine.Use(natsby.WithCustomRecovery(func(c *natsby.Context, err interface{}) {
		container.Logger.Error().
			Str("subject", c.Msg.Subject).
			Str("replyChan", c.Msg.Reply).
			Msg(fmt.Sprintf("%+v", err))

		if c.Msg.Reply != "" {
			c.Engine.NatsConnection.Publish(c.Msg.Reply, []byte("")) // TODO: Return an error object
		}
	}))

	SubscribeHandlers(container, engine)

	engine.Run(func() {
		container.Logger.Info().Msg("Ready to receive messages")
	})
}
