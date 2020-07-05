package app

import (
	"time"

	"github.com/nats-io/nats.go"
)

// WithLogger wraps handler with logs
func (service *Service) WithLogger(handler MsgHandler) MsgHandler {
	return func(m *nats.Msg) ([]byte, error) {
		service.Logger.Debug().
			Str("subject", m.Subject).
			Msg("received")

		start := time.Now()

		// TODO: Check for error?
		value, err := handler(m)

		end := time.Now()
		latency := end.Sub(start)

		if err != nil {
			service.Logger.Error().
				Str("subject", m.Subject).
				Err(err).
				Msg("")
		}

		service.Logger.Info().
			Str("subject", m.Subject).
			Dur("Latency", latency).
			Msg("complete")

		return value, err
	}
}

// WithResponse Checks for reply channel and sends response back
func (service *Service) WithResponse(handler MsgHandler) MsgHandler {
	return func(m *nats.Msg) ([]byte, error) {
		value, err := handler(m)

		if m.Reply != "" {
			service.Broker.Publish(m.Reply, value)
		}

		return value, err
	}
}
