package app

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// WithLogger wraps handler with logs
func (service *Service) WithLogger(topic string, handler HandlerFunc) (string, HandlerFunc) {
	return topic, func(m *nats.Msg) ([]byte, error) {
		service.Logger.Debug().
			Str("subject", m.Subject).
			Msg("received")

		start := time.Now()

		value, err := handler(m)

		end := time.Now()
		latency := end.Sub(start)

		if err != nil {
			service.Logger.Error().
				Str("subject", m.Subject).
				Err(err).
				Msg(fmt.Sprintf("%+v", err))
		}

		service.Logger.Info().
			Str("subject", m.Subject).
			Dur("latencyMS", latency).
			Str("replyChan", m.Reply).
			Msg("processed")

		return value, err
	}
}

// WithResponse Checks for reply channel and sends response back
func (service *Service) WithResponse(topic string, handler HandlerFunc) (string, HandlerFunc) {
	return topic, func(m *nats.Msg) ([]byte, error) {
		value, err := handler(m)

		if m.Reply != "" {
			service.Broker.Publish(m.Reply, value)
		}

		return value, err
	}
}
