package traffic

import (
	"bytes"
	"fmt"
	"testing"
)

// Must match JSON encoded version of "storeTests"
var jsonRef = `[
  {
    "_id": "6feff3dd9af0607e3006a71212a05792389b583b",
    "time": "2012-11-01T22:08:41Z",
    "remote": "81.2.69.160",
    "method": "GET",
    "uri": "/",
    "protocol": "HTTP/1.0",
    "status": 0,
    "payload_size": 0,
    "hour_of_day": 0,
    "remote_ip": "81.2.69.160",
    "country": "United Kingdom",
    "city": "London",
    "timezone": "Europe/London",
    "location": {
      "lat": 51.5142,
      "lon": -0.0931
    },
    "client_time": "2012-11-01T22:08:41Z"
  },
  {
    "_id": "ea6bf6d5fa0c42a787d041134aec8cf3f2a52842",
    "time": "0001-01-01T00:00:00Z",
    "remote": "",
    "method": "",
    "uri": "",
    "protocol": "",
    "status": 0,
    "payload_size": 0,
    "hour_of_day": 0
  }
]
`
// Must be the output of no output.
var jsonEmpty = `[
]
`

func TestJSONStore(t *testing.T) {
	var buf bytes.Buffer
	store, err := NewJSONStore(&buf)
	if err != nil {
		t.Fatal(err)
	}

	// add tests to the store
	addRequests(t, store, storeTests)

  if buf.String() != jsonRef {
	   fmt.Printf("var jsonRef = `%s`\n", buf.String())
     t.Fatal("JSON did not match reference")
  }

	buf.Reset()
	// Test closing without adding contents
	store, err = NewJSONStore(&buf)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Close()
	if err != nil {
		t.Fatal(err)
	}
	// Test we can call close multiple times.
	err = store.Close()
	if err != nil {
		t.Fatal(err)
	}
  if buf.String() != jsonEmpty {
	   fmt.Printf("var jsonRef = `%s`\n", buf.String())
     t.Fatal("JSON did not match reference")
  }
}
