package traffic

import "time"

// Request represents a single server request.
type Request struct {
	LocalTime  time.Time // Server local time of the request
	Remote       string    // Host or IP of the requester
	URI        string    // The requested URI
	StatusCode int       // The status code returned
	Payload    int       // The size of the returned body in bytes
}
