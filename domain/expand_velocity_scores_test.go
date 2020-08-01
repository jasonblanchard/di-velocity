package domain

import (
	"testing"
	"time"
)

func TestExpandVelocityScores(t *testing.T) {
	start := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2020, time.January, 31, 0, 0, 0, 0, time.UTC)
	velocities := DailyVelocities{
		{
			Day:       time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC),
			Score:     1,
			CreatorID: "123",
		},
		{
			Day:       time.Date(2020, time.January, 15, 0, 0, 0, 0, time.UTC),
			Score:     2,
			CreatorID: "123",
		},
		{
			Day:       time.Date(2020, time.January, 30, 0, 0, 0, 0, time.UTC),
			Score:     3,
			CreatorID: "123",
		},
	}

	output := ExpandVelicityScores(velocities, start, end)
	expectedLen := 31
	if len(output) != expectedLen {
		t.Errorf("length %v, wanted %v", len(output), expectedLen)
	}

	if output[0].Score != velocities[0].Score {
		t.Errorf("first score %v, wanted %v", output[0].Score, velocities[0].Score)
	}

	if output[14].Score != velocities[1].Score {
		t.Errorf("first score %v, wanted %v", output[14].Score, velocities[1].Score)
	}

	if output[29].Score != velocities[2].Score {
		t.Errorf("first score %v, wanted %v", output[29].Score, velocities[2].Score)
	}
}
