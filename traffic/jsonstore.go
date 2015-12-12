package traffic

import (
	"encoding/json"
	"fmt"
	"io"
)

// jsonStore writes an array of requests to
// the supplied writer
type jsonStore struct {
	out    io.Writer
	closed bool
	queued []byte
}

// NewJSONStore returns a RequestStore that writes an array of requests 
// marshalled as JSON to the supplied writer.
func NewJSONStore(out io.Writer) (RequestStore, error) {
	j := &jsonStore{out: out}
	fmt.Fprintln(j.out, "[")
	return j, nil
}

// Store a request.
// We keep one request, so we know if we should output a 
// separating comma.
func (j *jsonStore) Store(r Request) error {
	if j.queued != nil {
		fmt.Fprintln(j.out, "  " + string(j.queued) + ",")
	}
	var err error
	j.queued, err = json.MarshalIndent(r, "  ", "  ")
	return err
}

// Since we don't handle actual storage
// this only clears any queued objects.  
func (j *jsonStore) RemoveAll() error {
	j.queued = nil
	return nil
}

// Close will flush the remaining queue
// and close the array.
func (j *jsonStore) Close() error {
	if j.queued != nil {
		fmt.Fprintln(j.out, "  "+ string(j.queued))
		j.queued = nil
	}
	if !j.closed {
		fmt.Fprintln(j.out, "]")
		j.closed = true
	}
	return nil
}
