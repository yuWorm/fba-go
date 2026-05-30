package task

import "context"

type SchedulerService interface {
	Reload(ctx context.Context) error
	Execute(ctx context.Context, schedulerID int64) error
	Enable(ctx context.Context, schedulerID int64, enabled bool) error
}

type CompatibleStatus string

const (
	StatusPending CompatibleStatus = "PENDING"
	StatusStarted CompatibleStatus = "STARTED"
	StatusSuccess CompatibleStatus = "SUCCESS"
	StatusRetry   CompatibleStatus = "RETRY"
	StatusFailure CompatibleStatus = "FAILURE"
)

func MapAsynqState(state string) CompatibleStatus {
	switch state {
	case "PENDING":
		return StatusPending
	case "STARTED":
		return StatusStarted
	case "SUCCESS":
		return StatusSuccess
	case "RETRY":
		return StatusRetry
	case "FAILURE":
		return StatusFailure
	case "active":
		return StatusStarted
	case "completed":
		return StatusSuccess
	case "retry":
		return StatusRetry
	case "archived":
		return StatusFailure
	default:
		return StatusPending
	}
}
