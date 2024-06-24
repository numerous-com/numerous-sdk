package push

import (
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/gql/jobs"
)

func stopJobs(id string) error {
	jobsByTool, err := jobs.JobsByTool(id, gql.GetClient())
	if err != nil {
		return err
	}

	for _, job := range jobsByTool {
		_, err = jobs.JobStop(job.ID, gql.GetClient())
		if err != nil {
			return err
		}
	}

	return nil
}
