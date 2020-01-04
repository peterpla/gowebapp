package queue

import "github.com/peterpla/lead-expert/pkg/adding"

// ********** ********** ********** ********** ********** **********

// nullSystem implements Queue interface for a null in-memory queue.
// I.e., Create and Connect are null, Add's are thrown away.
type nullSystem struct {
	//
}

func NewNullQueue(qi *QueueInfo) Queue {
	q := &nullSystem{}
	if err := q.InfoFromConfig(qi); err != nil {
		return nil
	}
	return q
}

func (gct *nullSystem) Create(qi *QueueInfo) error {
	return nil // queue already created
}

func (gct *nullSystem) Connect(qi *QueueInfo) error {
	return nil
}

func (gct *nullSystem) Add(qi *QueueInfo, request *adding.Request) error {
	return nil
}

func (gct *nullSystem) InfoFromConfig(qi *QueueInfo) error {
	qi.Name = "null"
	qi.ServiceToHandle = "null"
	qi.HandlerEndpoint = "/task_handler"
	return nil
}
