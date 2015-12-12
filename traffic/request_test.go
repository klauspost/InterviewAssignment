package traffic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/google/gofuzz"
	"github.com/oschwald/geoip2-golang"
)

// Return a time no later than 5000 years from unix datum.
// JSON cannot handle dates after year 9999.
func fuzzTime(t *time.Time, c fuzz.Continue) {
	sec := c.Rand.Int63()
	nsec := c.Rand.Int63()
	// No more than 5000 years in the future
	sec %= 5000 * 365 * 24 * 60 * 60
	*t = time.Unix(sec, nsec)
}

// Test if hash generation is deterministic
// We do 10 runs with different data.
func TestGenerateHash(t *testing.T) {
	for i := int64(0); i < 10; i++ {
		var a, b Request

		// Fill requests with values
		f := fuzz.New()
		f.NumElements(0, 50) // Slioes have between 0 and 50 elements
		f.NilChance(0.1)     // Nilable types have 10% chance of being nil
		// Be sure we don't generate invalid times.
		f.Funcs(fuzzTime)
		f.RandSource(rand.New(rand.NewSource(i)))
		f.Fuzz(&a)
		a.GenerateHash()

		// Reset random seed
		f.RandSource(rand.New(rand.NewSource(i)))
		f.Fuzz(&b)
		b.GenerateHash()
		if a.ID == "" {
			t.Fatal("hash was not set")
		}
		if a.ID != b.ID {
			t.Fatalf("Hash was not deterministic, %q != %q", a.ID, b.ID)
		}
		if len(a.ID) != 20*2 {
			t.Fatalf("unexpected hash length, was %d, expected 40", len(a.ID))
		}
	}
}

type reqTest struct {
	in  Request
	out Request
}

var someTime, _ = time.Parse(time.RFC3339, "2012-11-01T22:08:41+00:00")

var reqTestsNoGeo = []reqTest{
	reqTest{Request{}, Request{}},
	// Test that valid IP addresses are transferred
	reqTest{
		in:  Request{ID: "ABCdefgf", ServerTime: someTime, Remote: "1.2.3.4", Method: "GET", URI: "/", Protocol: "HTTP/1.0", StatusCode: 0, Payload: 0, RemoteIP: "", Country: "", City: "", Timezone: "", Location: map[string]float64(nil), ClientTime: nil},
		out: Request{ID: "ABCdefgf", ServerTime: someTime, Remote: "1.2.3.4", Method: "GET", URI: "/", Protocol: "HTTP/1.0", StatusCode: 0, Payload: 0, RemoteIP: "1.2.3.4", Country: "", City: "", Timezone: "", Location: map[string]float64(nil), ClientTime: nil, HourOfDay: 22},
	},
	// Remote host names should not be transferred.
	reqTest{
		in:  Request{ID: "ABCdefgf", ServerTime: someTime, Remote: "peytz.dk", Method: "GET", URI: "/", Protocol: "HTTP/1.0", StatusCode: 0, Payload: 0, RemoteIP: "", Country: "", City: "", Timezone: "", Location: map[string]float64(nil), ClientTime: nil},
		out: Request{ID: "ABCdefgf", ServerTime: someTime, Remote: "peytz.dk", Method: "GET", URI: "/", Protocol: "HTTP/1.0", StatusCode: 0, Payload: 0, RemoteIP: "", Country: "", City: "", Timezone: "", Location: map[string]float64(nil), ClientTime: nil, HourOfDay: 22},
	},
}

func TestEnrichNoGeo(t *testing.T) {
	GeoDB = nil
	for i, test := range reqTestsNoGeo {
		req := test.in
		req.Enrich()

		err := compareReqJSON(req, test.out)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Enrich test %d PASSED.", i)
	}
}

var reqTestsGeo = []reqTest{
	reqTest{Request{}, Request{}},
	// Test that valid IP addresses are transferred
	reqTest{
		in:  Request{ID: "ABCdefgf", ServerTime: someTime, Remote: "1.2.3.4", Method: "GET", URI: "/", Protocol: "HTTP/1.0", StatusCode: 0, Payload: 0, RemoteIP: "", Country: "", City: "", Timezone: "", Location: map[string]float64(nil), ClientTime: nil},
		out: Request{ID: "ABCdefgf", ServerTime: someTime, Remote: "1.2.3.4", Method: "GET", URI: "/", Protocol: "HTTP/1.0", StatusCode: 0, Payload: 0, RemoteIP: "1.2.3.4", Country: "", City: "", Timezone: "", Location: map[string]float64(nil), ClientTime: nil, HourOfDay: 22},
	},
	// Remote host names should not be transferred.
	reqTest{
		in:  Request{ID: "ABCdefgf", ServerTime: someTime, Remote: "peytz.dk", Method: "GET", URI: "/", Protocol: "HTTP/1.0", StatusCode: 0, Payload: 0, RemoteIP: "", Country: "", City: "", Timezone: "", Location: map[string]float64(nil), ClientTime: nil},
		out: Request{ID: "ABCdefgf", ServerTime: someTime, Remote: "peytz.dk", Method: "GET", URI: "/", Protocol: "HTTP/1.0", StatusCode: 0, Payload: 0, RemoteIP: "", Country: "", City: "", Timezone: "", Location: map[string]float64(nil), ClientTime: nil, HourOfDay: 22},
	},
	// Test an IP that is in the sample database.
	reqTest{
		in:  Request{ID: "ABCdefgf", ServerTime: someTime, Remote: "81.2.69.160", Method: "GET", URI: "/", Protocol: "HTTP/1.0", StatusCode: 0, Payload: 0, RemoteIP: "", Country: "", City: "", Timezone: "", Location: map[string]float64(nil), ClientTime: nil},
		out: Request{ID: "ABCdefgf", ServerTime: someTime, Remote: "81.2.69.160", Method: "GET", URI: "/", Protocol: "HTTP/1.0", StatusCode: 0, Payload: 0, RemoteIP: "81.2.69.160", Country: "United Kingdom", City: "London", Timezone: "Europe/London", Location: map[string]float64{"lat": 51.5142, "lon": -0.0931}, ClientTime: &someTime, HourOfDay: 22},
	},
}

func TestEnrichGeoDB(t *testing.T) {
	var err error
	GeoDB, err = geoip2.Open("testdata/GeoIP2-City-Test.mmdb")
	if err != nil {
		t.Skip(err)
	}
	for i, test := range reqTestsGeo {
		req := test.in
		req.Enrich()

		err := compareReqJSON(req, test.out)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("Enrich test %d PASSED.", i)
	}
}

// Compare that a JSON marshalled version of two requests match.
func compareReqJSON(got, expect Request) error {
	gotJ, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		return err
	}

	expJ, err := json.MarshalIndent(expect, "", "  ")
	if err != nil {
		return err
	}

	if bytes.Compare(gotJ, expJ) != 0 {
		return fmt.Errorf("requests did not match.\n---Expected:\n%s\n---Got:\n%s", expJ, gotJ)
	}
	return nil
}
