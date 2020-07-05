package app

import (
	"time"

	"github.com/nats-io/nats.go"
)

// WithLogger wraps handler with logs
func (service *Service) WithLogger(handler nats.MsgHandler) nats.MsgHandler {
	return func(m *nats.Msg) {
		service.Logger.Debug().
			Str("subject", m.Subject).
			Msg("received")

		start := time.Now()

		// TODO: Check for error?
		handler(m)

		end := time.Now()
		latency := end.Sub(start)
		service.Logger.Info().
			Str("subject", m.Subject).
			Dur("Latency", latency).
			Msg("complete")
	}
}
