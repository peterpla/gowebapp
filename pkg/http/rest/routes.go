package rest

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/peterpla/gowebapp/pkg/adding"
)

func Routes(a adding.Service) http.Handler {
	// log.Printf("Routes - enter\n")
	router := httprouter.New()

	router.POST("/api/v1/requests", addRequest(a))

	// log.Printf("Routes - exit\n")
	return router
}

// addRequest returns a handler for POST /requests
func addRequest(s adding.Service) httprouter.Handle {
	// log.Printf("rest.AddRequest - enter/exit")
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// log.Printf("rest.AddRequest handler - enter\n")
		decoder := json.NewDecoder(r.Body)

		var newRequest adding.Request
		err := decoder.Decode(&newRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("rest.AddRequest handler - decoded request: %+v\n", newRequest)

		s.AddRequest(newRequest)

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode("New request added.")

		// log.Printf("rest.AddRequest handler - exit\n")
	}
}
