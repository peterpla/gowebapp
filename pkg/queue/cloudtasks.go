package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"

	"github.com/peterpla/lead-expert/pkg/config"
	"github.com/peterpla/lead-expert/pkg/request"
)

// ********** ********** ********** ********** ********** **********

// gctSystem implements QueueSystem specifically for Google Cloud Tasks
type gctSystem struct {
	//
}

func NewGCTQueue(qi *QueueInfo) Queue {
	q := &gctSystem{}
	if err := q.InfoFromConfig(qi); err != nil {
		return nil
	}
	return q
}

func (gct *gctSystem) Create(qi *QueueInfo) error {
	return nil // queue already created
}

func (gct *gctSystem) Connect(qi *QueueInfo) error {
	// connect to the Google Cloud Tasks queue
	// e.g., confirm it exists
	return nil
}

func (gct *gctSystem) Add(qi *QueueInfo, request *request.Request) error {
	// add the request to the GCT queue

	// JSON-encode the incoming req as the payload message
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("queue.AddToQueue: %v", err)
	}
	// log.Printf("%s.queue.AddRequest, queueName: %s, nextService: %s, requestJSON: %s\n",
	// 	serviceInfo.GetServiceName(), qi.Name, qi.ServiceToHandle, string(requestJSON))

	// Create a new Cloud Tasks client instance.
	// See https://godoc.org/cloud.google.com/go/cloudtasks/apiv2
	ctx := context.Background()
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("NewClient: %v", err)
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"

	// Build the Task payload.
	// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#CreateTaskRequest
	qReq := &taskspb.CreateTaskRequest{
		Parent: qi.Name,
		Task: &taskspb.Task{
			// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#AppEngineHttpRequest
			MessageType: &taskspb.Task_AppEngineHttpRequest{
				AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Headers:    headers,
					AppEngineRouting: &taskspb.AppEngineRouting{
						Service: qi.ServiceToHandle,
					},
					RelativeUri: qi.HandlerEndpoint,
					Body:        requestJSON,
				},
			},
		},
		ResponseView: taskspb.Task_FULL, // includes Body in response
	}

	createdTask, err := client.CreateTask(ctx, qReq)
	if err != nil {
		return fmt.Errorf("queue.AddRequest: %v", err)
	}

	// isolate TASK_ID, the last component of createdTask.name
	// returned Task struct: https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#Task
	i := strings.LastIndex(createdTask.Name, "/")
	taskID := createdTask.Name[i+1:]
	_ = taskID

	// note whether the requestJSON passed to CreateTaskRequest became the Body of the created task
	// log.Printf("%s.queue.Add, created task %s on queuePath %q: %+v, \n",
	// 	serviceInfo.GetServiceName(), taskID, qi.Name, createdTask)

	return nil
}

func (gct *gctSystem) InfoFromConfig(qi *QueueInfo) error {
	cfg := config.GetConfigPointer()

	qi.Name = fmt.Sprintf("projects/%s/locations/%s/queues/%s", cfg.ProjectID, cfg.StorageLocation, cfg.QueueName)
	qi.ServiceToHandle = cfg.NextServiceName
	qi.HandlerEndpoint = "/task_handler" // default endpoint for Google Cloud Tasks
	// log.Printf("cloudtasks.InfoFromConfig, QueueInfo: %+v\n", qi)

	return nil
}
