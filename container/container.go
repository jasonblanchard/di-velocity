package container

import (
	"os"
	"time"

	"github.com/jasonblanchard/di-velocity/repository"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

// Container holds all service dependencies
type Container struct {
	NATSConnection *nats.Conn
	Logger         *zerolog.Logger
	NATSQueue      string
	TestMode       bool
	Repository     repository.T
}

// Input arg for Service
type Input struct {
	PostgresUser     string
	PostgresPassword string
	PostgresDbName   string
	NatsURL          string
	Debug            bool
	Pretty           bool
	TestMode         bool
	NATSQueue        string
}

// New initializes service dependencies
func New(input *Input) (*Container, error) {
	container := &Container{
		NATSQueue: input.NATSQueue,
		TestMode:  input.TestMode,
	}

	repository, err := repository.NewPostgres(input.PostgresUser, input.PostgresPassword, input.PostgresDbName)
	if err != nil {
		return container, err
	}

	container.Repository = repository

	nc, err := initializeNATS(input.NatsURL)
	if err != nil {
		return container, err
	}

	container.NATSConnection = nc

	logger := initializeLogger(input.Debug, input.Pretty)
	container.Logger = logger

	return container, nil
}

func initializeLogger(debug, pretty bool) *zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	if pretty == true {
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug == true {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	zerolog.DurationFieldUnit = time.Second
	return &logger
}

// InitializeNATS sets up NATS connection
func initializeNATS(url string) (*nats.Conn, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return nc, nil
}

// InitExternalConfig pull in configuration from environment or file
func InitExternalConfig(path string) string {
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
