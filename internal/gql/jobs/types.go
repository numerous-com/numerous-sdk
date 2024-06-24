package jobs

type jobsByToolResponse struct {
	JobsByTool []Job
}

type Job struct {
	ID string
}

type jobStopResponse struct {
	Message string
}
