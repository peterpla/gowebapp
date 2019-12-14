package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/spf13/viper"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"

	"github.com/peterpla/lead-expert/pkg/adding"
	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

// Store data in Google Cloud Tasks queue
type GCT struct {
	// requests []adding.Request
}

// Add the request to the repository, i.e., Google Cloud Tasks queue
func (g *GCT) AddRequest(req adding.Request) error {
	// log.Printf("%s.queue.AddRequest - enter\n", serviceInfo.GetServiceName())

	// assemble the Task queue path components
	projectID := viper.GetString("ProjectID")
	locationID := viper.GetString("TasksLocation")
	queueName := serviceInfo.GetQueueName()
	serviceToHandleRequest := serviceInfo.GetNextServiceName()
	endpoint := "/task_handler"

	// JSON-encode the incoming req as the payload message
	requestJSON, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("queue.AddRequest: %v", err)
	}
	// log.Printf("queue.AddRequest, Body: %s\n", string(requestJSON))

	// taskID, err := g.AddToCloudTasksQ(projectID, locationID, queueName, serviceToHandleRequest, endpoint, requestJSON)
	_, err = g.AddToCloudTasksQ(projectID, locationID, queueName, serviceToHandleRequest, endpoint, requestJSON)
	if err != nil {
		return err
	}
	// log.Printf("queue.AddRequest - taskID %d created, exit\n", taskID)

	return nil
}

// AddToCloudTasksQ handles the Cloud Tasks-specifics to add to a queue
func (g *GCT) AddToCloudTasksQ(projectID, locationID, queueName, serviceToHandleRequest, handlerEndpoint string, requestJSON []byte) (taskID int, err error) {
	// log.Printf("%s.AddToCloudTasksQ entered, projectID: %q, locationID: %q, queueName: %q, serviceToHandleRequest: %q, handlerEndpoint: %q, requestJSON: %q\n",
	// 	serviceInfo.GetServiceName(), projectID, locationID, queueName, serviceToHandleRequest, handlerEndpoint, string(requestJSON))

	// Create a new Cloud Tasks client instance.
	// See https://godoc.org/cloud.google.com/go/cloudtasks/apiv2
	ctx := context.Background()
	client, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return 0, fmt.Errorf("NewClient: %v", err)
	}

	queuePath := fmt.Sprintf("projects/%s/locations/%s/queues/%s", projectID, locationID, queueName)
	// log.Printf("queue.AddToCloudTasksQ, queuePath: %q, service: %q\n", queuePath, serviceToHandleRequest)

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
						Service: serviceToHandleRequest,
					},
					RelativeUri: handlerEndpoint,
					Body:        requestJSON,
				},
			},
		},
	}

	createdTask, err := client.CreateTask(ctx, qReq)
	if err != nil {
		return 0, fmt.Errorf("queue.AddRequest: %v", err)
	}

	// isolate taskname (number), the last component of createdTask.name
	// returned Task struct: https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#Task
	i := strings.LastIndex(createdTask.Name, "/")
	taskID, err = strconv.Atoi(createdTask.Name[i+1:])
	if err != nil {
		return 0, fmt.Errorf("strconv.Atoi: %v", err)
	}

	sn := serviceInfo.GetServiceName()
	b := string(createdTask.GetAppEngineHttpRequest().GetBody())
	log.Printf("%s.queue.AddToCloudTasksQ, task %d created: %+v, on queuePath: %q, Body: %s",
		sn, taskID, createdTask, queuePath, b)

	return taskID, nil
}
