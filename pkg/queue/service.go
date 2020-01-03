package queue

import (
	"github.com/peterpla/lead-expert/pkg/adding/"
)

// QueueInfo identifies key properties of a queue
type QueueInfo struct {
	Name string // Name of queue
	ServiceToHandle string // Name of service to receive this request
}

// Queue is an abstract interface that defines operations
// required for any supported queueing system
type Queue interface {
	Create(q *QueueInfo) error
	Connect(q *QueueInfo) error
	Add(q *QueueInfo, request *adding.Request) error
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
	
	// JSON-encode the incoming req as the payload message
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("queue.AddToQueue: %v", err)
	}
	log.Printf("%s.queue.AddRequest, queueName: %s, nextService: %s, requestJSON: %s\n",
		serviceInfo.GetServiceName(), qi.Name, cfg.NextServiceName, string(requestJSON))

	return qs.queue.Add(qi, requestJSON)
}
