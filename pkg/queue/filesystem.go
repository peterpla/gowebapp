package queue

import "github.com/peterpla/lead-expert/pkg/request"

// ********** ********** ********** ********** ********** **********

// fileSystem implements Queue interface for a file system-based queue.
type fileSystem struct {
	//
}

func NewFileSystemQueue(qi *QueueInfo) Queue {
	q := &fileSystem{}
	if err := q.InfoFromConfig(qi); err != nil {
		return nil
	}
	return q
}

func (fs *fileSystem) Create(qi *QueueInfo) error {
	// create a directory named qi.Name in TempDir
	return nil
}

func (fs *fileSystem) Connect(qi *QueueInfo) error {
	return nil
}

func (fs *fileSystem) Add(qi *QueueInfo, request *request.Request) error {
	// marshall the request's JSON
	// write to a file named request.RequestID in the queue's directory qi.Name
	// POST to localhost:[port]/task_handler of the next service
	return nil
}

func (fs *fileSystem) InfoFromConfig(qi *QueueInfo) error {
	qi.Name = "null"
	qi.ServiceToHandle = "null"
	qi.HandlerEndpoint = "/task_handler"
	return nil
}
