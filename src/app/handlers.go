package app

import (
	"time"

	entryMessage "github.com/jasonblanchard/di-velocity/src/di_messages/entry"
	"github.com/jasonblanchard/di-velocity/src/di_messages/insights"
	insightsMessage "github.com/jasonblanchard/di-velocity/src/di_messages/insights"
	"github.com/jasonblanchard/di-velocity/src/op"
	"github.com/jasonblanchard/di-velocity/src/utils"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

// Handlers configures all the handlers
func (service *Service) Handlers() {
	if service.TestMode == true {
		service.Broker.QueueSubscribe("insights.store.drop", service.BrokerQueueName, service.WithLogger(service.HandleDrop()))
	}
	service.Broker.QueueSubscribe("info.entry.updated", service.BrokerQueueName, service.WithLogger(service.HandleEntryUpdated()))
	service.Broker.QueueSubscribe("insights.increment.dailyCounter", service.BrokerQueueName, service.WithLogger(service.HandleIncrementDailyCounter()))
	service.Broker.QueueSubscribe("insights.get.velocity", service.BrokerQueueName, service.WithLogger(service.handleGetVelocity()))
}

// HandleDrop handles drop
func (service *Service) HandleDrop() nats.MsgHandler {
	return func(m *nats.Msg) {
		err := op.DropDailyCounts(service.Store)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
		}

		service.Broker.Publish(m.Reply, []byte(""))
	}
}

// HandleEntryUpdated handles entry updated
func (service *Service) HandleEntryUpdated() nats.MsgHandler {
	return func(m *nats.Msg) {
		entryUpdatedMessage := &entryMessage.InfoEntryUpdated{}
		err := proto.Unmarshal(m.Data, entryUpdatedMessage)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
		}

		normalizedDay := utils.NormalizeTime(time.Unix(entryUpdatedMessage.Payload.UpdatedAt.Seconds, 0))
		day := utils.TimeToProtoTime(normalizedDay)

		incrementDailyCounterRequest := &insightsMessage.IncrementDailyCounter{
			Payload: &insightsMessage.IncrementDailyCounter_Payload{
				Day:       &day,
				CreatorId: entryUpdatedMessage.Payload.CreatorId,
			},
		}

		request, err := proto.Marshal(incrementDailyCounterRequest)

		if err != nil {
			utils.HandleMessageError(m.Subject, err)
		}

		service.Broker.Publish("insights.increment.dailyCounter", request)

		if m.Reply != "" {
			service.Broker.Publish(m.Reply, []byte(""))
		}
	}
}

// HandleIncrementDailyCounter handles incrementing counter for a day
func (service *Service) HandleIncrementDailyCounter() nats.MsgHandler {
	return func(m *nats.Msg) {
		requestMessage := &insights.IncrementDailyCounter{}
		err := proto.Unmarshal(m.Data, requestMessage)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
		}

		day := time.Unix(requestMessage.Payload.Day.Seconds, 0).UTC()

		err = op.IncrementDailyCounter(service.Store, day, requestMessage.Payload.CreatorId)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
		}
	}
}

// handleGetVelocity handles getting velocity scores
func (service *Service) handleGetVelocity() nats.MsgHandler {
	return func(m *nats.Msg) {
		requestMessage := &insightsMessage.GetVelocityRequest{}
		err := proto.Unmarshal(m.Data, requestMessage)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
			return
			// TODO: Respond with error type
		}

		normalizedStart := utils.NormalizeTime(time.Unix(requestMessage.Payload.Start.Seconds, 0).UTC())
		normalizedEnd := utils.NormalizeTime(time.Unix(requestMessage.Payload.End.Seconds, 0).UTC())

		dailyCounts, err := op.GetDailyCounts(service.Store, normalizedStart, normalizedEnd)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
			return
		}

		dailyVelocities := dailyCounts.ToVelocityScores()

		responseMessage := &insightsMessage.GetVelocityResponse{
			Payload: dailyVelocities.ToDtoPayload(),
		}

		response, err := proto.Marshal(responseMessage)
		if err != nil {
			utils.HandleMessageError(m.Subject, err)
		}

		service.Broker.Publish(m.Reply, response)
	}
}
