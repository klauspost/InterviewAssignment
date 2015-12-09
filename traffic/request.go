package traffic

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"time"
)

// Request represents a single server request.
type Request struct {
	ID         string    `json:"id,omitempty"`
	LocalTime  time.Time `json:"time"`         // Server local time of the request
	Remote     string    `json:"remote"`       // Host or IP of the requester
	URI        string    `json:"uri"`          // The requested URI
	StatusCode int       `json:"status"`       // The status code returned
	Payload    int       `json:"payload_size"` // The size of the returned body in bytes
}

// GenerateHash will generate a unique hash for a request
// and populate the ID field of the request.
// Hashes are based on the JSON representation of the request
// and will be deterministic for the same data.
func (r *Request) GenerateHash() {
	r.ID = ""

	// We use a JSON encoding of the request for our hash.
	// This has the advantage that we can hide fields from the hash.
	//
	// It also implies that if the structure is changed, the hash will change.
	b, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}

	hash := sha1.New()
	_, err = hash.Write(b)
	if err != nil {
		panic(err)
	}
	r.ID = hex.EncodeToString(hash.Sum(nil))
}

// RequestStore indicates an interface that can be used
// to store requests.
type RequestStore interface {
	Store(Request) error
	RemoveAll() error
	Close() error
}
