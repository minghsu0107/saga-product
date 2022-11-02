package pkg

import (
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
)

// Time2pbTimestamp converter
func Time2pbTimestamp(now time.Time) *timestamp.Timestamp {
	s := int64(now.Unix())
	n := int32(now.Nanosecond())
	return &timestamp.Timestamp{Seconds: s, Nanos: n}
}
