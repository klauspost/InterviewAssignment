package traffic

import (
	"fmt"
	"log"

	"gopkg.in/olivere/elastic.v3"
)

// elasticStore is an internal representation of
// a RequestStore.
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
//
// Requests are stored in bulk requests, so errors will be
// reported at a later point. That means that a request may not be
// stored even if the Store call did not return an error,
// and that any error returned are likely from a previous store.
//
// If an error has been encountered, the storer will not recover.
// If Close returns successfully all requests are stored successfully.
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

	// Create the template for new indexes
	err = e.createTemplate()
	if err != nil {
		return nil, err
	}

	// Start async saver
	go e.startSaver()
	return e, nil
}

// Store a request in elastic
//
// Requests are stored in bulk requests, so errors will be
// reported at a later point. That means that a request may not be
// stored even if nil is returned, and that any error returned are
// likely from a previous store.
func (e *elasticStore) Store(r Request) error {
	e.queue <- &r
	return e.err.Err()
}

// startSaver will start an async saver
func (e *elasticStore) startSaver() {
	// When this function returns, always close the finished channel
	defer close(e.finished)

	bulk := elastic.NewBulkService(e.client)
	for {
		// Get item of the queue
		r, ok := <-e.queue

		// If channel was closed, store the rest and return
		if !ok {
			if bulk.NumberOfActions() == 0 {
				return
			}
			res, err := bulk.Do()
			e.err.Set(err)
			if res != nil && res.Errors {
				e.err.Set(fmt.Errorf("bulk index has error. %d failed, %d succeeded", len(res.Failed()), len(res.Succeeded())))
			}
			return
		}
		// Remove ID, ES has that as a separate field
		id := r.ID
		r.ID = ""

		// Get destination index based on the Request
		index := r.Index(e.index)

		// Create the request and add it to the bulk saver
		req := elastic.NewBulkIndexRequest().Index(index).Type("request").Id(id).Doc(r)
		bulk.Add(req)

		// If we have collected 500 documents, send the request.
		if bulk.NumberOfActions() >= 500 {
			// BulkService.Do() resets the request, so we can reuse it.
			res, err := bulk.Do()
			if err != nil {
				e.err.Set(err)
				return
			}
			if res.Errors {
				e.err.Set(fmt.Errorf("bulk index returned error(s). %d failed, %d succeeded", len(res.Failed()), len(res.Succeeded())))
			}
		}
	}
}

// createTemplate will create/update a template for new indexes
// See https://www.elastic.co/guide/en/elasticsearch/guide/current/index-templates.html
func (e elasticStore) createTemplate() error {
	t := map[string]interface{}{
		"template": e.index + "-*",
		"order":    1,
		"settings": map[string]interface{}{
			"number_of_shards": 1,
		},
		"mappings": map[string]interface{}{
			"request": map[string]interface{}{
				"properties": map[string]interface{}{
					"time": map[string]interface{}{
						"type": "date",
					},
					"remote": map[string]interface{}{
						"type":  "string",
						"index": "not_analyzed",
					},
					"remote_ip": map[string]interface{}{
						"type": "ip",
					},
					"uri": map[string]interface{}{
						"type":  "string",
						"index": "not_analyzed",
					},
					"method": map[string]interface{}{
						"type":  "string",
						"index": "not_analyzed",
					},
					"protocol": map[string]interface{}{
						"type":  "string",
						"index": "not_analyzed",
					},
					"country": map[string]interface{}{
						"type":  "string",
						"index": "not_analyzed",
					},
					"city": map[string]interface{}{
						"type":  "string",
						"index": "not_analyzed",
					},
					"timezone": map[string]interface{}{
						"type":  "string",
						"index": "not_analyzed",
					},
					"location": map[string]interface{}{
						"type": "geo_point",
						"fielddata": map[string]interface{}{
							"format":    "compressed",
							"precision": "1km",
						},
					},
				},
			},
		},
	}
	_, err := e.client.IndexPutTemplate(e.index).BodyJson(&t).Do()
	return err
}

// RemoveAll will remove all contents of the indexes.
func (e *elasticStore) RemoveAll() error {
	// Get all indexes starting with the index prefix.
	res, err := e.client.IndexGet(e.index + "-*").AllowNoIndices(true).Do()
	if err != nil {
		return err
	}

	// Delete all indexes.
	for k := range res {
		log.Println("Deleting", k)
		ci, err := e.client.DeleteIndex(k).Do()
		if err != nil {
			return err
		}
		if !ci.Acknowledged {
			return fmt.Errorf("elastic did not acknowledge index deletion")
		}
	}
	return nil
}

// Close will flush the remaining stores and
// return any errors that was encountered.
func (e *elasticStore) Close() error {
	select {
	case <-e.finished:
		return e.err.Err()
	default:
	}
	close(e.queue)
	<-e.finished
	return e.err.Err()
}
