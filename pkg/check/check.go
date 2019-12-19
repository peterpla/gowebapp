// Check package provides consistency checks to detect anomalies early
package check

import (
	"errors"

	"github.com/google/uuid"

	"github.com/peterpla/lead-expert/pkg/adding"
)

// ErrZeroUUID - zero-valued UUID
var ErrZeroUUID = errors.New("zero-valued UUID")

func RequestID(req adding.Request) error {
	zeroUUID := uuid.Nil
	if req.RequestID == zeroUUID {
		return ErrZeroUUID
	}
	return nil
}
