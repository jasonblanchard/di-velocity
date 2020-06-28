package main

import (
	"testing"

	nats "github.com/nats-io/nats.go"
	. "github.com/smartystreets/goconvey/convey"
)

func TestIntegration(t *testing.T) {
	Convey("The system", t, func() {
		nc, err := nats.Connect(nats.DefaultURL)
		if err != nil {
			panic(err)
		}

		Convey("works", func() {
			err := nc.Publish("insights.get.velocity", []byte("testing, testing"))
			So(err, ShouldBeNil)
			So(true, ShouldEqual, true)
		})
	})
}
