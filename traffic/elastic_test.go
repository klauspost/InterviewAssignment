package traffic

import (
	"encoding/json"
	"flag"
	"os"
	"testing"

	"gopkg.in/olivere/elastic.v3"
)

var elasticHost = flag.String("elastic", "http://127.0.0.1:9200", "url to elasticseach server (http)")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

var storeTests = []Request{
	// Test a populated struct
	Request{ID: "ABCdefgf", ServerTime: someTime, Remote: "81.2.69.160", Method: "GET", URI: "/", Protocol: "HTTP/1.0", StatusCode: 0, Payload: 0, RemoteIP: "81.2.69.160", Country: "United Kingdom", City: "London", Timezone: "Europe/London", Location: map[string]float64{"lat": 51.5142, "lon": -0.0931}, ClientTime: &someTime},
	// Test an empty struct
	Request{},
}

// Add all requests to the store.
// The test will fail if any error is encountered.
// The RequestStore is closed before returning.
func addRequests(t *testing.T, s RequestStore, r []Request) {
	for i, req := range r {
		req.GenerateHash()
		err := s.Store(req)
		if err != nil {
			t.Fatalf("test item %d returned: %s", i, err.Error())
		}
	}
	err := s.Close()
	if err != nil {
		t.Fatal("closing store returned:", err)
	}
}

func TestElastic(t *testing.T) {
	testIndex := "es-test-index"
	store, err := NewElastic(*elasticHost, testIndex)
	if err != nil {
		t.Skip("Unable to connect to elasticsearch server.\nUse -elastic parameter to set server address.")
	}

	// Remove all content when we exit
	defer func() {
		err := store.RemoveAll()
		if err != nil {
			t.Fatal(err)
		}
	}()

	// add tests to the store
	addRequests(t, store, storeTests)

	// Test if we can read them back
	client, err := elastic.NewClient(elastic.SetURL(*elasticHost))
	if err != nil {
		t.Fatal(err)
	}

	for i, req := range storeTests {
		req.GenerateHash()
		res, err := client.Get().Index(req.Index(testIndex)).Id(req.ID).FetchSource(true).Do()
		if err != nil {
			t.Fatal("test", i, "error:", err)
		}
		if !res.Found {
			t.Fatal("expected to find a document")
		}
		if res.Type != "request" {
			t.Fatal("expected type request, got", res.Type)
		}
		var got Request
		err = json.Unmarshal(*res.Source, &got)
		if err != nil {
			t.Fatal(err)
		}
		got.GenerateHash()
		compareReqJSON(req, got)
		t.Logf("Elastic store #%d PASSED", i)
	}

	// Test RemoveAll
	store, err = NewElastic(*elasticHost, testIndex)
	if err != nil {
		t.Fatal(err)
	}
	store.RemoveAll()
	for i, req := range storeTests {
		req.GenerateHash()
		_, err := client.Get().Index(req.Index(testIndex)).Id(req.ID).Do()
		if err == nil {
			t.Fatal("deleted test", i, "found index.")
		}
		t.Logf("Elastic delete #%d PASSED", i)
	}
	err = store.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Test closing without adding contents
	store, err = NewElastic(*elasticHost, testIndex)
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
}
