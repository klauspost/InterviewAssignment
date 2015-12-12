package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/klauspost/InterviewAssignment/traffic"
)

var updateGolden = flag.Bool("update", false, "update golden reference files")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

// Tests import of a small file and compares it to a reference
// JSON encoding.
// Use `-update` when the model is updated to update the reference file.
func TestImport(t *testing.T) {
	logOut = ioutil.Discard

	// Output to an internal buffer
	var buf bytes.Buffer
	store, err := traffic.NewJSONStore(&buf)
	if err != nil {
		t.Fatal(err)
	}
	err = importFile("testdata/sample-log.txt.gz", store)
	if err != nil {
		t.Fatal(err)
	}
	err = store.Close()
	if err != nil {
		t.Fatal(err)
	}

	var ref, got []map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &ref)
	if err != nil {
		t.Fatal(err)
	}

	if *updateGolden {
		err := ioutil.WriteFile("testdata/sample-log.txt.ref", buf.Bytes(), 0666)
		if err != nil {
			t.Fatal(err)
		}
		t.Skip("golden files updated")
	}

	// Open and decode the reference file.
	reference, err := ioutil.ReadFile("testdata/sample-log.txt.ref")
	if err != nil {
		t.Fatal(err)
	}

	err = json.Unmarshal(reference, &got)
	if err != nil {
		t.Fatal(err)
	}

	// Convert to comparable maps
	iref := indexMaps(t, ref)
	igot := indexMaps(t, got)
	if !reflect.DeepEqual(iref, igot) {
		for k, v := range iref {
			t.Logf("REF:%q:%v\n", k, v)
			t.Logf("GOT:%q:%v\n", k, igot[k])
		}
		t.Fatal("import did not yield reference values")
	}
}

// indexMaps will convert a slice of elements to an indexed map, where index is "_id".
// This allows us to compare content independent of order.
func indexMaps(t *testing.T, in []map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{})
	for i, v := range in {
		idi, ok := v["_id"]
		if !ok {
			t.Fatalf("could not find _id for item %d", i)
		}
		id, ok := idi.(string)
		if !ok {
			t.Fatalf("_id was not a string for item %d", i)
		}
		dst[id] = v
	}
	return dst
}
