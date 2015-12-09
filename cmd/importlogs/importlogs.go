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

// Log format, see https://github.com/satyrius/gonx#format
var format = flag.String("format", `$remote_addr - - [$time_local] "$request" $status $size`, "Log format")

var timeFormat = flag.String("timeformat", `02/Jan/2006:15:04:05 -0700`, "Time format in Go time.Parse format.")

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
			panic(err)
		}
	}
}

// Import a single file.
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
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		req, err := parseRecord(rec)
		if err != nil {
			return err
		}
		fmt.Printf("%#v\n", *req)
	}
	return nil
}

// parseRecord parses a single entry and returns a typed Request.
// TODO: Determine which errors should be returned.
func parseRecord(rec *gonx.Entry) (*traffic.Request, error) {
	// Convert each record to a request object
	var req traffic.Request
	f, err := rec.Field("time_local")
	if err == nil {
		t, err := time.Parse(*timeFormat, f)
		if err != nil {
			return nil, err
		}
		req.LocalTime = t
	}
	req.Remote, _ = rec.Field("remote_addr")
	req.URI, _ = rec.Field("request")
	f, err = rec.Field("status")
	if err == nil {
		req.StatusCode, _ = strconv.Atoi(f)
	}
	f, err = rec.Field("size")
	if err == nil {
		req.Payload, _ = strconv.Atoi(f)
	}
	return &req, nil
}
