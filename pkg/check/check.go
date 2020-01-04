// Check package provides consistency checks to detect anomalies early
package check

import (
	"errors"

	"github.com/google/uuid"
	"github.com/peterpla/lead-expert/pkg/request"
)

// ErrZeroUUID - zero-valued UUID
var ErrZeroUUID = errors.New("zero-valued UUID")

func RequestID(req request.Request) error {
	zeroUUID := uuid.Nil
	if req.RequestID == zeroUUID {
		return ErrZeroUUID
	}
	return nil
}
