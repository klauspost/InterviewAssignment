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
	"os"
	"strconv"
	"time"

	"github.com/klauspost/InterviewAssignment/traffic"
	"github.com/klauspost/pgzip"
	"github.com/satyrius/gonx"
)

// Parameters
var (
	// Log format, see https://github.com/satyrius/gonx#format
	format        = flag.String("format", `$remote_addr - - [$time_local] "$request" $status $size`, "Log format")
	timeFormat    = flag.String("timeformat", `02/Jan/2006:15:04:05 -0700`, "Time format in Go time.Parse format.")
	continueError = flag.Bool("e", false, "continnue to next file if an error occurs")
)

// Local variables
var (
	exitCode = 0
)

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
	for _, file := range args {
		err := importFile(file)
		if err != nil {
			report(file, err)
		}
	}
	os.Exit(exitCode)
}

// Report an error, and exit depending on the 'continueError'
func report(file string, err error) {
	if file == "" {
		fmt.Fprint(os.Stderr, err)
	} else {
		fmt.Fprint(os.Stderr, file+": "+err.Error())
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
func importFile(file string) error {
	fi, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	gr, err := pgzip.NewReader(br)
	if err != nil {
		return err
	}
	defer gr.Close()

	reader := gonx.NewReader(gr, *format)
	n := 0
	start := time.Now()
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		req, err := parseEntry(rec)
		if err != nil {
			return err
		}
		_ = req
		n++
	}
	elapsed := time.Since(start)
	fmt.Printf("Processing %q took %s, processing %d entries.\n", file, elapsed, n)
	fmt.Printf("%0.2f entries/sec.", float64(n)/elapsed.Seconds())
	return nil
}

// parseEntry parses a single entry and returns a typed Request.
// Individual fields that are missing are ignored, but if a field is found
// it must be parseable, otherwise the error will be returned.
func parseEntry(rec *gonx.Entry) (*traffic.Request, error) {
	// Convert each record to a request object
	var req traffic.Request

	// Individual fields that are missing are ignored.
	req.Remote, _ = rec.Field("remote_addr")
	req.URI, _ = rec.Field("request")

	f, err := rec.Field("time_local")
	if err == nil {
		t, err := time.Parse(*timeFormat, f)
		if err != nil {
			return nil, err
		}
		req.LocalTime = t
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
