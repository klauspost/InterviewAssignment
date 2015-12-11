// importslogs imports and indexes apache style log files.
//
// Reference logs are available here: http://ita.ee.lbl.gov/html/contrib/NASA-HTTP.html
//
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/klauspost/InterviewAssignment/traffic"
	"github.com/klauspost/pgzip"
	"github.com/oschwald/geoip2-golang"
	"github.com/satyrius/gonx"
)

// Executable flags.
// See doc.go for more details.
var (
	format        = flag.String("format", `$remote_addr - - [$time_local] "$method $uri $protocol" $status $size`, "Log format")
	timeFormat    = flag.String("timeformat", `02/Jan/2006:15:04:05 -0700`, "Time format in Go time.Parse format.")
	continueError = flag.Bool("e", false, "continnue to next file if an error occurs")
	elasticHost   = flag.String("elastic", "http://127.0.0.1:9200", "url to elasticseach server (http)")
	clean         = flag.Bool("clean", false, "clean the index before adding content")
	geoDB         = flag.String("geodb", "", "MaxMind GeoLite2 or GeoIP2 mmdb database to translate IP to location")
)

// Local variables.
var (
	exitCode = 0 // Exitcode. Used if 'continueError' is set.
)

// Print usage help and exit with exit code 2
func usage() {
	fmt.Fprintln(os.Stderr, "usage: importlogs [flags] file1.gz [file2.gz...]")
	fmt.Fprintln(os.Stderr, "\tImports gzipped log files.")
	fmt.Fprintln(os.Stderr, "flags:")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		usage()
	}

	esEnv := os.Getenv("ELASTICSEARCH_PORT_9200_TCP")

	if esEnv != "" {
		esEnv = strings.Replace(esEnv, "tcp://", "http://", 1)
		log.Println("Using ELASTICSEARCH_PORT_9200_TCP environment variable:", esEnv)
		elasticHost = &esEnv
	}

	log.Println("Connecting to host:", *elasticHost)

	// Be sure we exit with the right exitcode.
	defer func() {
		os.Exit(exitCode)
	}()

	// Create an elasticsearch storer.
	store, err := traffic.NewElastic(*elasticHost, "requests")
	failOnErr(err)

	// Be sure we close the store and return the proper exitcode
	defer func() {
		err := store.Close()
		failOnErr(err)
	}()

	// Load GeoIP database
	if *geoDB != "" {
		traffic.GeoDB, err = geoip2.Open(*geoDB)
		failOnErr(err)
	}

	// Clean the database if requested
	if *clean {
		err := store.RemoveAll()
		failOnErr(err)
	}

	// Open all input files
	for _, file := range args {
		err := importFile(file, store)
		if err != nil {
			report(file, err)
		}
	}
}

// Report an error and always fail
func failOnErr(err error) {
	if err == nil {
		return
	}
	report("", err)
	os.Exit(2)
}

// Report an error, and exit depending on the 'continueError'
func report(file string, err error) {
	if file == "" {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Fprintln(os.Stderr, file+": "+err.Error())
	}
	// Exit now
	if !*continueError {
		os.Exit(2)
	}
	// Report error at end
	exitCode = 2
}

// importFile will Import a single file.
// The file is assumed to be gzipped.
func importFile(file string, store traffic.RequestStore) error {
	// Open file
	fi, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fi.Close()

	// Unzip the input stream
	br := bufio.NewReader(fi)
	gr, err := pgzip.NewReader(br)
	if err != nil {
		return err
	}
	defer gr.Close()

	// Track metrics for this file
	n := 0
	start := time.Now()

	// Use gonx to split log files
	reader := gonx.NewReader(gr, *format)
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// Parse the entry
		req, err := parseEntry(rec)
		if err != nil {
			return err
		}
		// We have an entry. Generate a hash for it, and enrich it.
		req.GenerateHash()
		req.Enrich()

		// Send it to the store
		err = store.Store(*req)
		if err != nil {
			return err
		}

		// Report metrics
		n++
		if n%1000 == 0 {
			elapsed := time.Since(start)
			fmt.Printf("Processed %d, %0.2f entries/sec.\n", n, float64(n)/elapsed.Seconds())
		}
	}
	// Print overall metrics
	elapsed := time.Since(start)
	fmt.Printf("Processing %q took %s, processing %d entries.\n", file, elapsed, n)
	fmt.Printf("%0.2f entries/sec.", float64(n)/elapsed.Seconds())
	return nil
}

// parseEntry parses a single entry and returns a typed Request.
// Individual fields that are missing are ignored, but if a field is found
// it must be parseable, otherwise an error will be returned.
func parseEntry(rec *gonx.Entry) (*traffic.Request, error) {
	// Convert each record to a request object
	var req traffic.Request

	// Individual fields that are missing are ignored.
	req.Remote, _ = rec.Field("remote_addr")
	req.URI, _ = rec.Field("uri")
	req.Method, _ = rec.Field("method")
	req.Protocol, _ = rec.Field("protocol")

	f, err := rec.Field("time_local")
	if err == nil {
		t, err := time.Parse(*timeFormat, f)
		if err != nil {
			return nil, err
		}
		req.ServerTime = t
	}

	f, err = rec.Field("status")
	if err == nil {
		req.StatusCode, err = strconv.Atoi(f)
		if err != nil {
			return nil, err
		}
	}

	// Size can be "-" on bodyless responses
	f, err = rec.Field("size")
	if err == nil && f != "-" {
		req.Payload, err = strconv.Atoi(f)
		if err != nil {
			return nil, err
		}
	}
	return &req, nil
}
