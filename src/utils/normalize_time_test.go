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

	Convey("works when we go in and out of proto time", t, func() {
		originalTime := time.Date(2020, time.January, 1, 10, 2, 03, 04, time.UTC)
		protoTime := TimeToProtoTime(originalTime)

		timeFromProto := time.Unix(protoTime.Seconds, 0)

		normalized := NormalizeTime(timeFromProto)
		So(normalized.String(), ShouldEqual, "2020-01-01 00:00:00 +0000 UTC")
	})
}
