package worker

import (
	"github.com/golang/protobuf/proto"
	"github.com/guilhermehubner/worker/broker"
	"golang.org/x/net/context"
)

type jobInfoKey struct{}

type GenFunc func(proto.Message) error

type JobHandle func(context.Context, GenFunc) error

// JobType settings of job which should be passed to RegisterJob
type JobType struct {
	Name     string    // The queue name
	Handle   JobHandle // The job handler function
	Priority uint      // Priority from 1 to 10000
	Retry    uint8     // Retry count 1 to 255
}

type JobInfo struct {
	MessageID string
	Name      string
	Priority  uint
	Retry     uint8
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

func injectJobInfo(ctx context.Context, job JobType, message *broker.Message) context.Context {
	ctx = context.WithValue(ctx, jobInfoKey{}, JobInfo{
		MessageID: message.ID(),
		Name:      job.Name,
		Priority:  job.Priority,
		Retry:     job.Retry,
	})

	return ctx
}

func JobInfoFromContext(ctx context.Context) JobInfo {
	return ctx.Value(jobInfoKey{}).(JobInfo)
}
