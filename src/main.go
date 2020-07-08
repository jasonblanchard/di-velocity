package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/jasonblanchard/di-velocity/src/app"
	"github.com/spf13/viper"

	_ "github.com/lib/pq"

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

	serviceInput := &app.ServiceInput{
		PostgresUser:     viper.GetString("db_user"),
		PostgresPassword: viper.GetString("db_password"),
		PostgresDbName:   viper.GetString("db_name"),
		NatsURL:          nats.DefaultURL,
		Debug:            viper.GetBool("debug"),
		Pretty:           viper.GetBool("pretty"),
		TestMode:         viper.GetBool("test_mode"),
	}

	service, err := app.NewService(serviceInput)
	if err != nil {
		fmt.Printf("Cannot configure application: %s", err)
		os.Exit(1)
	}

	service.Use(service.WithLogger)
	service.Handlers()

	if configFile != "" {
		service.Logger.Info().Msg(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	go func() {
		// Wait for signal
		<-c
		service.Store.Close()
		service.Broker.Drain()
		os.Exit(0)
	}()

	service.Logger.Info().Msg("Ready to receive messages")
	runtime.Goexit()
}
