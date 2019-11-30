package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/spf13/viper"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"

	"github.com/peterpla/gowebapp/pkg/adding"
)

// Store data in Google Cloud Tasks queue
type GCT struct {
	// requests []adding.Request
}

// Add the request to the repository, i.e., Google Cloud Tasks queue
func (g *GCT) AddRequest(req adding.Request) error {
	// log.Printf("queue.AddRequest - enter\n")

	// Build the Task queue path.
	projectID := viper.GetString("ProjectID")
	locationID := viper.GetString("TasksLocation")
	queueID := viper.GetString("TaskInitialRequestQ")

	// JSON-encode the incoming req as the payload message
	requestJSON, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("queue.AddRequest: %v", err)
	}
	// log.Printf("queue.AddRequest, Body: %q\n", requestJSON)

	taskID, err := g.AddToCloudTasksQ(projectID, locationID, queueID, "InitialRequest", "/task_handler", requestJSON)
	if err != nil {
		return err
	}
	log.Printf("queue.AddRequest - taskID %d created, exit\n", taskID)

	return nil
}

// AddToCloudTasksQ handles the Cloud Tasks-specifics to add to a queue
func (g *GCT) AddToCloudTasksQ(projectID, locationID, queueName, serviceName, handlerEndpoint string, requestJSON []byte) (taskID int, err error) {
	// Create a new Cloud Tasks client instance.
	// See https://godoc.org/cloud.google.com/go/cloudtasks/apiv2
	ctx := context.Background()
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return 0, fmt.Errorf("NewClient: %v", err)
	}

	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s",
		projectID, locationID, queueName)
	// log.Printf("queue.AddToCloudTasksQ, queuePath: %q\n", queuePath)

	// Build the Task payload.
	// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#CreateTaskRequest
	qReq := &taskspb.CreateTaskRequest{
		Parent: queuePath,
		Task: &taskspb.Task{
			// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#AppEngineHttpRequest
			MessageType: &taskspb.Task_AppEngineHttpRequest{
				AppEngineHttpRequest: &taskspb.AppEngineHttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					AppEngineRouting: &taskspb.AppEngineRouting{
						Service: "InitialRequest",
					},
					RelativeUri: "/task_handler",
					Body:        requestJSON,
				},
			},
		},
	}

	createdTask, err := client.CreateTask(ctx, qReq)
	if err != nil {
		return 0, fmt.Errorf("queue.AddRequest: %v", err)
	}
	// TODO: isolate taskname (number) at end of createdTask.name (path)
	taskID, err = strconv.Atoi("012345")
	if err != nil {
		return 0, err
	}

	log.Printf("queue.AddRequest, added to %s\n... createdTask: %+v\n... Body: %s",
		queuePath, createdTask, createdTask.GetAppEngineHttpRequest().GetBody())

	return taskID, nil
}
