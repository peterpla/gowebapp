package appengine

import (
	"log"
	"net/http"

	"github.com/peterpla/lead-expert/pkg/serviceInfo"
)

func getAppEngineInfo(w http.ResponseWriter, r *http.Request) (taskName, queueName string) {
	sn := serviceInfo.GetNextServiceName()
	// var taskName string
	t, ok := r.Header["X-Appengine-Taskname"]
	if !ok || len(t[0]) == 0 {
		// You may use the presence of the X-Appengine-Taskname header to validate
		// the request comes from Cloud Tasks.
		log.Printf("%s Invalid Task: No X-Appengine-Taskname request header found\n", sn)
		http.Error(w, "Bad Request - Invalid Task", http.StatusBadRequest)
		return
	}
	taskName = t[0]

	// Pull useful headers from Task request.
	q, ok := r.Header["X-Appengine-Queuename"]
	queueName = ""
	if ok {
		queueName = q[0]
	}

	return taskName, queueName
}
