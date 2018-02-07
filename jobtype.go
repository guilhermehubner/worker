package worker

import (
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
)

type JobHandle func(context.Context, func(proto.Message) error) error

// JobType settings of job which should be passed to RegisterJob
type JobType struct {
	Name     string    // The queue name
	Handle   JobHandle // The job handler function
	Priority uint      // Priority from 1 to 10000
	Retry    uint8     // Retry count 1 to 255
}

type jobTypes []JobType

func (slice jobTypes) Len() int {
	return len(slice)
}

func (slice jobTypes) Less(i, j int) bool {
	return slice[i].Priority > slice[j].Priority
}

func (slice jobTypes) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}
