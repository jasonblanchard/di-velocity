package utils

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNormalizeTime(t *testing.T) {
	Convey("NormalizeTime", t, func() {
		initialTime, err := time.Parse(time.RFC3339, "2020-01-01T3:04:05+06:07")
		if err != nil {
			panic(err)
		}
		normalized := NormalizeTime(initialTime)
		normalizedIsoTime := normalized.Format(time.RFC3339)
		So(normalizedIsoTime, ShouldEqual, "2020-01-01T00:00:00Z")
	})
}
