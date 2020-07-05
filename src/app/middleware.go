package app

import "github.com/nats-io/nats.go"

// WithLogger wraps handler with logs
func (service *Service) WithLogger(handler nats.MsgHandler) nats.MsgHandler {
	return func(m *nats.Msg) {
		handler(m)
	}
}
