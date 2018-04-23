package renderings

type HealthCheckResponse struct {
	Message   string `json:"message"`
	CommitID  string `json:"commit_id"`
	BuildTime string `json:"build_time"`
}
