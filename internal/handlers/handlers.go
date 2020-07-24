package handlers

import (
	"time"

	"github.com/jasonblanchard/di-velocity/internal/container"
	entryMessage "github.com/jasonblanchard/di-velocity/internal/di_messages/entry"
	"github.com/jasonblanchard/di-velocity/internal/di_messages/insights"
	insightsMessage "github.com/jasonblanchard/di-velocity/internal/di_messages/insights"
	"github.com/jasonblanchard/di-velocity/internal/domain"
	"github.com/jasonblanchard/di-velocity/internal/mappers"
	"github.com/jasonblanchard/di-velocity/internal/repository"
	"github.com/jasonblanchard/natsby"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func withRepository(repository repository.T) natsby.HandlerFunc {
	return func(c *natsby.Context) {
		c.Set("repository", repository)
	}
}

// Subscribe configures all the handlers
func Subscribe(c *container.Container, e *natsby.Engine) {

	if c.TestMode == true {
		e.Subscribe("insights.store.drop", natsby.WithByteReply(), withRepository(c.Repository), handleDrop())
	}
	e.Subscribe("info.entry.updated", handleEntryUpdated())
	e.Subscribe("insights.increment.dailyCounter", natsby.WithByteReply(), withRepository(c.Repository), handleIncrementDailyCounter())
	e.Subscribe("insights.get.velocity", natsby.WithByteReply(), withRepository(c.Repository), handleGetVelocity())
}

func handleDrop() natsby.HandlerFunc {
	return func(c *natsby.Context) {
		repository, ok := c.Get("repository").(repository.T)

		if ok != true {
			c.Err = errors.Wrap(errors.New("cast error"), "Repository not initialized")
		}

		err := repository.DropDailyCounts()
		if err != nil {
			c.Err = errors.Wrap(err, "DropDailyCounts failed")
			return
		}

		c.ByteReplyPayload = []byte("")
	}
}

func handleEntryUpdated() natsby.HandlerFunc {
	return func(c *natsby.Context) {
		entryUpdatedMessage := &entryMessage.InfoEntryUpdated{}
		err := proto.Unmarshal(c.Msg.Data, entryUpdatedMessage)
		if err != nil {
			c.Err = errors.Wrap(err, "Unmarshall failed")
			return
		}

		normalizedDay := domain.NormalizeTime(time.Unix(entryUpdatedMessage.Payload.UpdatedAt.Seconds, 0))
		day := mappers.TimeToProtoTime(normalizedDay)

		incrementDailyCounterRequest := &insightsMessage.IncrementDailyCounter{
			Payload: &insightsMessage.IncrementDailyCounter_Payload{
				Day:       &day,
				CreatorId: entryUpdatedMessage.Payload.CreatorId,
			},
		}

		request, err := proto.Marshal(incrementDailyCounterRequest)

		if err != nil {
			c.Err = errors.Wrap(err, "Marshal failed")
			return
		}

		c.Engine.NatsConnection.Publish("insights.increment.dailyCounter", request)
	}
}

func handleIncrementDailyCounter() natsby.HandlerFunc {
	return func(c *natsby.Context) {
		repository, ok := c.Get("repository").(repository.T)

		if ok != true {
			c.Err = errors.Wrap(errors.New("cast error"), "Repository not initialized")
		}

		requestMessage := &insights.IncrementDailyCounter{}
		err := proto.Unmarshal(c.Msg.Data, requestMessage)
		if err != nil {
			c.Err = errors.Wrap(err, "unmarshall failed")
			return
		}

		day := time.Unix(requestMessage.Payload.Day.Seconds, 0).UTC()

		err = repository.IncrementDailyCounter(day, requestMessage.Payload.CreatorId)
		if err != nil {
			c.Err = errors.Wrap(err, "increment failed")
			return
		}
	}
}

func handleGetVelocity() natsby.HandlerFunc {
	return func(c *natsby.Context) {
		repository, ok := c.Get("repository").(repository.T)

		if ok != true {
			c.Err = errors.Wrap(errors.New("cast error"), "Repository not initialized")
		}

		requestMessage := &insightsMessage.GetVelocityRequest{}
		err := proto.Unmarshal(c.Msg.Data, requestMessage)
		if err != nil {
			// TODO: Respond with error type
			c.Err = errors.Wrap(err, "unmarshall failed")
			return
		}

		normalizedStart := domain.NormalizeTime(time.Unix(requestMessage.Payload.Start.Seconds, 0).UTC())
		normalizedEnd := domain.NormalizeTime(time.Unix(requestMessage.Payload.End.Seconds, 0).UTC())

		dailyCounts, err := repository.GetDailyCounts(normalizedStart, normalizedEnd)
		if err != nil {
			c.Err = errors.Wrap(err, "get daily counts failed")
			return
		}

		dailyVelocities := dailyCounts.ToVelocityScores()

		responseMessage := &insightsMessage.GetVelocityResponse{
			Payload: mappers.VelocitiesToProtoPayload(dailyVelocities),
		}

		message, err := proto.Marshal(responseMessage)
		if err != nil {
			c.Err = errors.Wrap(err, "Marshal failed")
			return
		}

		c.ByteReplyPayload = message
	}
}