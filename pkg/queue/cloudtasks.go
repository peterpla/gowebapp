package queue

import (
	"context"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"

	"github.com/peterpla/lead-experts/pkg/adding"
)

// ********** ********** ********** ********** ********** **********
// GCTAdapter implements QueueSystem specifically for Google Cloud Tasks
type gctSystem struct {
	//
}

func NewGCTQueue() Queue {
	return &gctSystem{}
}

func (gct *gctSystem) Create(qi *QueueInfo) error {
	return nil // queue already created
}

func (gct *gctSystem) Connect(qi *QueueInfo) error {
	// connect to the Google Cloud Tasks queue
	// e.g., confirm it exists
	return nil
}

func (gct *gctSystem) Add(qi *QueueInfo, request *adding.Request) error {
	// add the request to the GCT queue


	// Create a new Cloud Tasks client instance.
	// See https://godoc.org/cloud.google.com/go/cloudtasks/apiv2
	ctx := context.Background()
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("NewClient: %v", err)
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
						Service: serviceToHandleRequest,
					},
					RelativeUri: handlerEndpoint,
					Body:        requestJSON,
				},
			},
		},
		ResponseView: taskspb.Task_FULL, // includes Body in response
	}

	createdTask, err := client.CreateTask(ctx, qReq)
	if err != nil {
		return "", fmt.Errorf("queue.AddRequest: %v", err)
	}

	// isolate TASK_ID, the last component of createdTask.name
	// returned Task struct: https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#Task
	i := strings.LastIndex(createdTask.Name, "/")
	taskID = createdTask.Name[i+1:]

	// note whether the requestJSON passed to CreateTaskRequest became the Body of the created task
	// log.Printf("%s.queue.AddToCloudTasksQ, task %s created: %+v, on queuePath: %q\n",
	// 	serviceInfo.GetServiceName(), taskID, createdTask, queuePath)

	return taskID, nil

	return nil
}
