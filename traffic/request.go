package traffic

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net"
	"time"

	"github.com/oschwald/geoip2-golang"
	"gopkg.in/olivere/elastic.v3"
)

// GeoDB should be initialized to look up geolocation for raw IP addresses
// when enriching data.
var GeoDB *geoip2.Reader

// Request represents a single server request.
type Request struct {
	ID         string    `json:"ID,omitempty"`
	ServerTime time.Time `json:"time"`         // Server local time of the request
	Remote     string    `json:"remote"`       // Host or IP of the requester
	Method     string    `json:"method"`       // Request method used.
	URI        string    `json:"uri"`          // The requested URI
	Protocol   string    `json:"protocol"`     // Request protocol used.
	StatusCode int       `json:"status"`       // The status code returned
	Payload    int       `json:"payload_size"` // The size of the returned body in bytes

	// Enriched fields:
	RemoteIP   string             `json:"remote_ip,omitempty"`   // IP of the requester
	Country    string             `json:"country,omitempty"`     // Country of the requester
	City       string             `json:"city,omitempty"`        // City of the requester
	Timezone   string             `json:"timezone,omitempty"`    // Timezone of the requester
	Location   map[string]float64 `json:"location,omitempty"`    // GeoIP location.
	ClientTime *time.Time         `json:"client_time,omitempty"` // Time converted to the client timezone
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

// Enrich the Request data.
//
// This will attempt to derive as much information about the request
// as possible.
//
// If GeoDB has been populated, it will attempt to attach a location to the request.
func (r *Request) Enrich() {
	ip := net.ParseIP(r.Remote)
	if ip != nil {
		r.RemoteIP = r.Remote
	} else {
		// TODO: Look up IP from hostname
		// Use "public_suffix_list.dat" to determine TLD
	}

	// If we have a GeoDB and an IP address, try to fill in information.
	if GeoDB != nil && ip != nil {
		result, err := GeoDB.City(ip)
		if err == nil {
			r.City, _ = result.City.Names["en"]
			r.Country, _ = result.Country.Names["en"]
			r.Timezone = result.Location.TimeZone
			r.Location = elastic.GeoPointFromLatLon(result.Location.Latitude, result.Location.Longitude).Source()
			goloc, err := time.LoadLocation(r.Timezone)
			if err == nil {
				t := r.ServerTime.In(goloc)
				r.ClientTime = &t
			}
		}
	}
}

// Index returns an index based on a base name
// combined with the UTC date. This corresponds to
// a typical Logstash-type index name.
func (r Request) Index(base string) string {
	suffix := r.ServerTime.UTC().Format("2006.01.02")
	return base + "-" + suffix
}

// RequestStore indicates an interface that can be used
// to store requests.
type RequestStore interface {
	// Store a request in the backend.
	// Errors may be lazily reported depending
	// on the implementation
	Store(Request) error

	// RemoveAll must remove all previous data from the store.
	// The function must not return before the operation has been completed.
	// If an error is encountered some indexes may still remain.
	RemoveAll() error

	// Close the RequestStore.
	// If nil is returned the implementation must have saved all
	// requests when the function returns.
	Close() error
}
