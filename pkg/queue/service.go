package queue

import (
	"github.com/peterpla/lead-expert/pkg/adding"
)

// QueueInfo identifies key properties of a queue
type QueueInfo struct {
	Name            string // Name of queue
	ServiceToHandle string // Name of service to receive this request
	HandlerEndpoint string // Endpoint to receive this request
}

// Queue is an abstract interface that defines operations
// required for any supported queueing system
type Queue interface {
	Create(q *QueueInfo) error
	Connect(q *QueueInfo) error
	Add(q *QueueInfo, request *adding.Request) error
	InfoFromConfig(q *QueueInfo) error // populate QueueInfo with config
}

// ********** ********** ********** ********** ********** **********

// QueueService defines the business logic to interact with a queue,
// most of which pass through to the underlying adapter
type QueueService interface {
	CreateQueue(q *QueueInfo) error
	ConnectToQueue(q *QueueInfo) error
	AddToQueue(q *QueueInfo, request *adding.Request) error
}

type queueService struct {
	queue Queue // ???
}

func NewService(q Queue) QueueService {
	return &queueService{
		q,
	}
}

func (qs *queueService) CreateQueue(qi *QueueInfo) error {
	// initialize
	return qs.queue.Create(qi)
}

func (qs *queueService) ConnectToQueue(qi *QueueInfo) error {
	// initialize
	return qs.queue.Connect(qi)
}

func (qs *queueService) AddToQueue(qi *QueueInfo, request *adding.Request) error {
	return qs.queue.Add(qi, request)
}
