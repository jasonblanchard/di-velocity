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
		service.RegisterHandler("insights.store.drop", service.WithResponse(service.HandleDrop()))
	}
	service.RegisterHandler("info.entry.updated", service.HandleEntryUpdated())
	service.RegisterHandler("insights.increment.dailyCounter", service.HandleIncrementDailyCounter())
	service.RegisterHandler("insights.get.velocity", service.WithResponse(service.handleGetVelocity()))
}

// HandleDrop handles drop
func (service *Service) HandleDrop() MsgHandler {
	return func(m *nats.Msg) ([]byte, error) {
		err := op.DropDailyCounts(service.Store)
		if err != nil {
			return nil, err
		}

		return []byte(""), nil
	}
}

// HandleEntryUpdated handles entry updated
func (service *Service) HandleEntryUpdated() MsgHandler {
	return func(m *nats.Msg) ([]byte, error) {
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
			return nil, err
		}

		service.Broker.Publish("insights.increment.dailyCounter", request)

		return nil, nil
	}
}

// HandleIncrementDailyCounter handles incrementing counter for a day
func (service *Service) HandleIncrementDailyCounter() MsgHandler {
	return func(m *nats.Msg) ([]byte, error) {
		requestMessage := &insights.IncrementDailyCounter{}
		err := proto.Unmarshal(m.Data, requestMessage)
		if err != nil {
			return nil, err
		}

		day := time.Unix(requestMessage.Payload.Day.Seconds, 0).UTC()

		err = op.IncrementDailyCounter(service.Store, day, requestMessage.Payload.CreatorId)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}
}

// handleGetVelocity handles getting velocity scores
func (service *Service) handleGetVelocity() MsgHandler {
	return func(m *nats.Msg) ([]byte, error) {
		requestMessage := &insightsMessage.GetVelocityRequest{}
		err := proto.Unmarshal(m.Data, requestMessage)
		if err != nil {
			// TODO: Respond with error type
			return nil, err
		}

		normalizedStart := utils.NormalizeTime(time.Unix(requestMessage.Payload.Start.Seconds, 0).UTC())
		normalizedEnd := utils.NormalizeTime(time.Unix(requestMessage.Payload.End.Seconds, 0).UTC())

		dailyCounts, err := op.GetDailyCounts(service.Store, normalizedStart, normalizedEnd)
		if err != nil {
			return nil, err
		}

		dailyVelocities := dailyCounts.ToVelocityScores()

		responseMessage := &insightsMessage.GetVelocityResponse{
			Payload: dailyVelocities.ToDtoPayload(),
		}

		return proto.Marshal(responseMessage)
	}
}
