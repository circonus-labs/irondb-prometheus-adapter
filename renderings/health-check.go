package renderings

// HealthCheckResponse - the response structure for the HealthCheck handler
// which includes a message, the build's commit_id and the build's build_time
type HealthCheckResponse struct {
	Message   string `json:"message"`
	CommitID  string `json:"commit_id"`
	BuildTime string `json:"build_time"`
}
