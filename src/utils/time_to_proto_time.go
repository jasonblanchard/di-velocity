package utils

import (
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
)

// TimeToProtoTime converts Go time objec to protobuf time object
func TimeToProtoTime(time time.Time) timestamp.Timestamp {
	return timestamp.Timestamp{
		Seconds: time.Unix(),
	}
}
