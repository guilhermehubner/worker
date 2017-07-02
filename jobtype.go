package worker

import "golang.org/x/net/context"

type jobHandle func(context.Context, []byte) error

type JobType struct {
	Name     string
	Handle   jobHandle
	Priority uint
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
