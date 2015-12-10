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

// Initialize to look up geolocation for raw IP addresses
var GeoDB *geoip2.Reader

// Request represents a single server request.
type Request struct {
	ID         string    `json:"ID,omitempty"`
	LocalTime  time.Time `json:"time"`         // Server local time of the request
	Remote     string    `json:"remote"`       // Host or IP of the requester
	Method     string    `json:"method"`       // Request method used.
	URI        string    `json:"uri"`          // The requested URI
	Protocol   string    `json:"protocol"`     // Request protocol used.
	StatusCode int       `json:"status"`       // The status code returned
	Payload    int       `json:"payload_size"` // The size of the returned body in bytes

	// Enriched fields:
	RemoteIP string             `json:"remote_ip,omitempty"` // IP of the requester
	Country  string             `json:"country,omitempty"`
	City     string             `json:"city,omitempty"`
	Timezone string             `json:"timezone,omitempty"`
	Location map[string]float64 `json:"location,omitempty"`
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
// If GeoDB has been populated, it will attempt to attach a location to the request.
func (r *Request) Enrich() {
	ip := net.ParseIP(r.Remote)
	if ip != nil {
		r.RemoteIP = r.Remote
	} else {
		// TODO: Look up IP from hostname
		// Use "public_suffix_list.dat" to determine TLD
	}

	if GeoDB != nil && ip != nil {
		result, err := GeoDB.City(ip)
		if err == nil {
			r.City, _ = result.City.Names["en"]
			r.Country, _ = result.Country.Names["en"]
			r.Timezone = result.Location.TimeZone
			r.Location = elastic.GeoPointFromLatLon(result.Location.Latitude, result.Location.Longitude).Source()
		}
	}
}

// Index returns an index based on a base name
// combined with the UTC date. This corresponds to
// a typical Logstash-type index name.
func (r Request) Index(base string) string {
	suffix := r.LocalTime.UTC().Format("2006.01.02")
	return base + "-" + suffix
}

// RequestStore indicates an interface that can be used
// to store requests.
type RequestStore interface {
	Store(Request) error
	RemoveAll() error
	Close() error
}
