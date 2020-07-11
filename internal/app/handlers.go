package app

import (
	"time"

	entryMessage "github.com/jasonblanchard/di-velocity/internal/di_messages/entry"
	"github.com/jasonblanchard/di-velocity/internal/di_messages/insights"
	insightsMessage "github.com/jasonblanchard/di-velocity/internal/di_messages/insights"
	"github.com/jasonblanchard/di-velocity/internal/domain"
	"github.com/jasonblanchard/di-velocity/internal/repository"
	"github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// Handlers configures all the handlers
func (service *Service) Handlers() {
	if service.TestMode == true {
		service.RegisterHandler(service.WithResponse(service.handleDrop()))
	}
	service.RegisterHandler(service.handleEntryUpdated())
	service.RegisterHandler(service.handleIncrementDailyCounter())
	service.RegisterHandler(service.WithResponse(service.handleGetVelocity()))
}

func (service *Service) handleDrop() (string, HandlerFunc) {
	return "insights.store.drop", func(m *nats.Msg) ([]byte, error) {
		err := repository.DropDailyCounts(service.Store)
		if err != nil {
			return nil, errors.Wrap(err, "DropDailyCounts failed")
		}

		return []byte(""), nil
	}
}

func (service *Service) handleEntryUpdated() (string, HandlerFunc) {
	return "info.entry.updated", func(m *nats.Msg) ([]byte, error) {
		entryUpdatedMessage := &entryMessage.InfoEntryUpdated{}
		err := proto.Unmarshal(m.Data, entryUpdatedMessage)
		if err != nil {
			return nil, errors.Wrap(err, "Unmarshall failed")
		}

		normalizedDay := domain.NormalizeTime(time.Unix(entryUpdatedMessage.Payload.UpdatedAt.Seconds, 0))
		day := TimeToProtoTime(normalizedDay)

		incrementDailyCounterRequest := &insightsMessage.IncrementDailyCounter{
			Payload: &insightsMessage.IncrementDailyCounter_Payload{
				Day:       &day,
				CreatorId: entryUpdatedMessage.Payload.CreatorId,
			},
		}

		request, err := proto.Marshal(incrementDailyCounterRequest)

		if err != nil {
			return nil, errors.Wrap(err, "Marshal failed")
		}

		service.Broker.Publish("insights.increment.dailyCounter", request)

		return nil, nil
	}
}

func (service *Service) handleIncrementDailyCounter() (string, HandlerFunc) {
	return "insights.increment.dailyCounter", func(m *nats.Msg) ([]byte, error) {
		requestMessage := &insights.IncrementDailyCounter{}
		err := proto.Unmarshal(m.Data, requestMessage)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshall failed")
		}

		day := time.Unix(requestMessage.Payload.Day.Seconds, 0).UTC()

		err = repository.IncrementDailyCounter(service.Store, day, requestMessage.Payload.CreatorId)
		if err != nil {
			return nil, errors.Wrap(err, "increment failed")
		}

		return nil, nil
	}
}

func (service *Service) handleGetVelocity() (string, HandlerFunc) {
	return "insights.get.velocity", func(m *nats.Msg) ([]byte, error) {
		requestMessage := &insightsMessage.GetVelocityRequest{}
		err := proto.Unmarshal(m.Data, requestMessage)
		if err != nil {
			// TODO: Respond with error type
			return nil, errors.Wrap(err, "unmarshall failed")
		}

		normalizedStart := domain.NormalizeTime(time.Unix(requestMessage.Payload.Start.Seconds, 0).UTC())
		normalizedEnd := domain.NormalizeTime(time.Unix(requestMessage.Payload.End.Seconds, 0).UTC())

		dailyCounts, err := repository.GetDailyCounts(service.Store, normalizedStart, normalizedEnd)
		if err != nil {
			return nil, errors.Wrap(err, "get daily counts failed")
		}

		dailyVelocities := dailyCounts.ToVelocityScores()

		responseMessage := &insightsMessage.GetVelocityResponse{
			Payload: VelocitiesToProtoPayload(dailyVelocities),
		}

		return proto.Marshal(responseMessage)
	}
}
