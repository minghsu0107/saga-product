package pkg

import (
	"errors"

	"github.com/sony/sonyflake"
)

// IDGenerator is the inteface for generatring unique ID
type IDGenerator interface {
	NextID() (uint64, error)
}

func NewSonyFlake() (IDGenerator, error) {
	var st sonyflake.Settings
	sf := sonyflake.NewSonyflake(st)
	if sf == nil {
		return nil, errors.New("sonyflake not created")
	}
	return sf, nil
}
