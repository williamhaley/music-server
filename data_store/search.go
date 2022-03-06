package data_store

import (
	"sync"

	"github.com/meilisearch/meilisearch-go"
	log "github.com/sirupsen/logrus"
)

type SearchClient struct {
	client    *meilisearch.Client
	batches   map[string][]interface{}
	batchSize int
}

func NewSearchClient(meilisearchAddress string) *SearchClient {
	return &SearchClient{
		// client: meilisearch.NewClient(meilisearch.ClientConfig{
		// 	Host: meilisearchAddress,
		// }),
		batches:   map[string][]interface{}{},
		batchSize: 1000,
	}
}

func (s *SearchClient) AddDocument(index string, document interface{}) error {
	return nil

	s.batches[index] = append(s.batches[index], document)
	if len(s.batches[index]) == s.batchSize {
		if _, err := s.client.Index(index).AddDocuments(s.batches[index]); err != nil {
			return err
		}

		s.batches[index] = make([]interface{}, 0)
	}

	return nil
}

func (s *SearchClient) FlushBatches() error {
	for index, batch := range s.batches {
		if len(batch) > 0 {
			if _, err := s.client.Index(index).AddDocuments(s.batches[index]); err != nil {
				return err
			}
			s.batches[index] = make([]interface{}, 0)
		}
	}

	return nil
}

func (s *SearchClient) Search(searchTerm string) (interface{}, error) {
	indexes := []string{"track", "album", "artist"}

	wg := &sync.WaitGroup{}
	wg.Add(len(indexes))

	results := make(map[string]interface{})

	for _, index := range indexes {
		go func(index, searchTerm string) {
			if resp, err := s.client.Index(index).Search(searchTerm, &meilisearch.SearchRequest{}); err != nil {
				log.WithError(err).Errorf("error searching index %q", index)
			} else {
				results[index] = resp
			}

			wg.Done()
		}(index, searchTerm)
	}

	wg.Wait()

	return results, nil
}

func (s *SearchClient) DropAll() {
	// indexes := []string{"track", "album", "artist"}

	// for _, index := range indexes {
	// 	_, err := s.client.DeleteIndexIfExists(index)
	// 	if err != nil {
	// 		log.WithError(err).Errorf("error deleting index %q", index)
	// 	}
	// }
}
