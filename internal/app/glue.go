package app

import "github.com/nats-io/nats.go"

// HandlerFunc message handler function. Meant to be chained with middleware.
type HandlerFunc func(*nats.Msg) ([]byte, error)

// MessageHandlerChainToNatsHandler takes message handler and makes it compatible with Nats
func (service *Service) MessageHandlerChainToNatsHandler(handler HandlerFunc) nats.MsgHandler {
	return func(m *nats.Msg) {
		handler(m)
	}
}

// MiddlewareFunc accepts a topic and handler function, wraps it and returns a new handler func
type MiddlewareFunc func(string, HandlerFunc) (string, HandlerFunc)

// Use add global middleware
func (service *Service) Use(m MiddlewareFunc) {
	service.GlobalMiddleware = append(service.GlobalMiddleware, m)
}

// RegisterHandler wraps handler in default middleware and listens on Nats queue
func (service *Service) RegisterHandler(topic string, handler HandlerFunc) {
	handlerChain := handler
	for _, middleware := range service.GlobalMiddleware {
		_, handlerChain = middleware(topic, handlerChain)
	}
	wrappedHandler := service.MessageHandlerChainToNatsHandler(handlerChain)
	service.Broker.QueueSubscribe(topic, service.BrokerQueueName, wrappedHandler)
}
