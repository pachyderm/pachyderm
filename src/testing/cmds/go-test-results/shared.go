package gotestresults

import (
	"os"
	"time"
)

var PostgresqlUser = os.Getenv("POSTGRESQL_USER")
var PostgresqlPassword = os.Getenv("POSTGRESQL_PASSWORD")
var PostgresqlHost = os.Getenv("POSTGRESQL_HOST")

type JobInfo struct {
	Id              string    `json:"id,omitempty"`
	WorkflowId      string    `json:"workflow_id"`
	JobId           string    `json:"job_id"`
	JobName         string    `json:"job_name"`
	JobTimestamp    time.Time `json:"job_timestamp"`
	JobNumExecutors int       `json:"job_num_executors"`
	JobExecutor     int       `json:"job_executor"`
	Commit          string    `json:"commit"`
	Branch          string    `json:"branch"`
	Username        string    `json:"username"`
	Tag             string    `json:"tag"`
	PullRequests    string    `json:"pull_requests"`
}
