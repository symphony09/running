package running

import (
	"errors"
)

var (
	ErrPlanNotFound = errors.New("plan not found")

	ErrBuildWorkerFailed = errors.New("build worker failed")

	ErrWorkerPanic = errors.New("worker panic")
)
