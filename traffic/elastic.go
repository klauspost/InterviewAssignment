package traffic

import (
	"fmt"

	"gopkg.in/olivere/elastic.v3"
)

type elasticStore struct {
	client *elastic.Client
	index  string

	// Used for the async saver
	queue    chan *Request
	finished chan struct{}
	err      *syncErr
}

// NewElastic will return a RequestStore, that will
// send all requests to an elastic backend.
func NewElastic(host, index string) (RequestStore, error) {
	e := &elasticStore{index: index, err: &syncErr{}}
	e.queue = make(chan *Request, 1000)
	e.finished = make(chan struct{}, 0)

	// Create elastic client
	var err error
	e.client, err = elastic.NewClient(elastic.SetURL(host))
	if err != nil {
		return nil, err
	}

	// Create index, if it does not exist
	exists, err := e.client.IndexExists(e.index).Do()
	if err != nil {
		return nil, err
	}
	if !exists {
		// Index does not exist yet.
		ci, err := e.client.CreateIndex(e.index).Do()
		if err != nil {
			return nil, err
		}
		if !ci.Acknowledged {
			return nil, fmt.Errorf("elastic did not acknowledge index creation")
		}
	}
	// Start async saver
	go e.startSaver()
	return e, nil
}

// Store a request in elastic
func (e *elasticStore) Store(r Request) error {
	e.queue <- &r
	return e.err.Err()
}

// startSaver will start an async saver
func (e *elasticStore) startSaver() {
	defer close(e.finished)
	bulk := elastic.NewBulkService(e.client)
	for {
		// Get item of the queue
		r, ok := <-e.queue
		if !ok {
			_, err := bulk.Do()
			e.err.Set(err)
			return
		}
		id := r.ID
		r.ID = ""
		req := elastic.NewBulkIndexRequest().Index(e.index).Type("request").Id(id).Doc(r)
		bulk.Add(req)

		// If we have collected 500 documents, send the request.
		if bulk.NumberOfActions() >= 500 {
			// BulkService.Do() resets the request, so we can reuse it.
			_, err := bulk.Do()
			if err != nil {
				e.err.Set(err)
				return
			}
		}
	}
}

// RemoveAll all contents of the index.
func (e *elasticStore) RemoveAll() error {
	// Create index, if it does not exist
	exists, err := e.client.IndexExists(e.index).Do()
	if err != nil {
		return err
	}
	if exists {
		// Index does not exist yet.
		ci, err := e.client.DeleteIndex(e.index).Do()
		if err != nil {
			return err
		}
		if !ci.Acknowledged {
			return fmt.Errorf("elastic did not acknowledge index deletion")
		}
	}

	// Index does not exist yet.
	ci, err := e.client.CreateIndex(e.index).Do()
	if err != nil {
		return err
	}
	if !ci.Acknowledged {
		return fmt.Errorf("elastic did not acknowledge index creation")
	}
	return nil
}

// RemoveAll all contents of the index.
func (e *elasticStore) Close() error {
	close(e.queue)
	<-e.finished
	return nil
}
