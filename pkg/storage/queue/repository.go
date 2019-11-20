package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/spf13/viper"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"

	"github.com/peterpla/gowebapp/pkg/adding"
)

// Store data in Google Cloud Tasks queue
type GCT struct {
	requests []adding.Request
}

// Add saves the request to the repository
func (m *GCT) AddRequest(req adding.Request) error {
	log.Printf("cloudtasks.AddRequest - enter\n")

	// Create a new Cloud Tasks client instance.
	// See https://godoc.org/cloud.google.com/go/cloudtasks/apiv2
	ctx := context.Background()
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("NewClient: %v", err)
	}

	// Build the Task queue path.
	projectID := viper.Get("projectID")
	locationID := viper.Get("tasksLocation")
	queueID := viper.Get("tasksQRequests")

	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s",
		projectID, locationID, queueID)
	log.Printf("queue.AddRequest, queuePath: %q\n", queuePath)

	// JSON-encode the incoming req as the payload message
	message, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("queue.AddRequest: %v", err)
	}
	log.Printf("queue.AddRequest, Body: %q\n", message)

	// Build the Task payload.
	// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#CreateTaskRequest
	qReq := &taskspb.CreateTaskRequest{
		Parent: queuePath,
		Task: &taskspb.Task{
			// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#AppEngineHttpRequest
			MessageType: &taskspb.Task_AppEngineHttpRequest{
				AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
					HttpMethod:  taskspb.HttpMethod_POST,
					RelativeUri: "/task_handler",
					Body:        message,
				},
			},
		},
	}

	createdTask, err := client.CreateTask(ctx, qReq)
	if err != nil {
		return fmt.Errorf("queue.AddRequest: %v", err)
	}
	log.Printf("queue.AddRequest, createdTask: %+v\n", createdTask)

	log.Printf("cloudtasks.AddRequest - exit\n")

	return nil
}
