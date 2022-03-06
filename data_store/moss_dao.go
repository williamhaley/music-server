package data_store

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/couchbase/moss"
	log "github.com/sirupsen/logrus"
)

type MossDao struct {
	store      *moss.Store
	collection moss.Collection
}

func md5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func NewMossDao(persistenceDirectory string) *MossDao {
	store, collection, err := moss.OpenStoreCollection(persistenceDirectory, moss.StoreOptions{}, moss.StorePersistOptions{})
	if err != nil {
		log.WithError(err).Fatal("error creating collection")
	}

	return &MossDao{
		store:      store,
		collection: collection,
	}
}

func (m *MossDao) AddTrack(album *Album, artist *Artist, track *Track) error {
	batch, err := m.collection.NewBatch(0, 0)
	if err != nil {
		return err
	}
	defer batch.Close()

	if err := m.addRecord(fmt.Sprintf("track:by-id:%s", track.Id), track); err != nil {
		return err
	}

	if album != nil {
		trackSortKey := fmt.Sprintf("%06d:%s", track.TrackNumber, track.Id)
		if err := m.addRecord(fmt.Sprintf("track:for-album-id:%s:%s", album.Id, trackSortKey), track); err != nil {
			return err
		}
		if err := m.addRecord(fmt.Sprintf("album:by-name:%s", strings.ToLower(album.Name)), album); err != nil {
			return err
		}
		if err := m.addRecord(fmt.Sprintf("album:by-id:%s", album.Id), album); err != nil {
			return err
		}
	}

	if artist != nil {
		if album != nil {
			if err := m.addRecord(fmt.Sprintf("album:for-artist-id:%s:%s", artist.Id, album.Id), album); err != nil {
				return err
			}
		}
		if err := m.addRecord(fmt.Sprintf("artist:by-name:%s", strings.ToLower(artist.Name)), artist); err != nil {
			return err
		}
		if err := m.addRecord(fmt.Sprintf("artist:by-id:%s", artist.Id), artist); err != nil {
			return err
		}
	}

	if err = m.collection.ExecuteBatch(batch, moss.WriteOptions{}); err != nil {
		return err
	}

	return nil
}

func (m *MossDao) addRecord(key string, record interface{}) error {
	batch, err := m.collection.NewBatch(0, 0)
	if err != nil {
		return err
	}
	defer batch.Close()

	var writer = new(bytes.Buffer)
	if err := json.NewEncoder(writer).Encode(record); err != nil {
		return err
	}

	batch.Set([]byte(key), writer.Bytes())

	if err = m.collection.ExecuteBatch(batch, moss.WriteOptions{}); err != nil {
		return err
	}

	return nil
}

func (m *MossDao) GetAlbums() ([]*Album, error) {
	var albums []*Album
	if err := m.getRecords("album:by-name:", &albums); err != nil {
		return nil, err
	}
	return albums, nil
}

func (m *MossDao) GetAlbum(id string) (*Album, error) {
	var album Album
	if err := m.getRecord(fmt.Sprintf("album:by-id:%s", id), &album); err != nil {
		return nil, err
	}
	var tracks []*Track
	if err := m.getRecords(fmt.Sprintf("track:for-album-id:%s:", album.Id), &tracks); err != nil {
		return nil, err
	}
	album.Tracks = tracks
	return &album, nil
}

func (m *MossDao) GetArtists() ([]*Artist, error) {
	var artists []*Artist
	if err := m.getRecords("artist:by-name:", &artists); err != nil {
		return nil, err
	}
	return artists, nil
}

func (m *MossDao) GetArtist(id string) (*Artist, error) {
	var artist Artist
	if err := m.getRecord(fmt.Sprintf("artist:by-id:%s", id), &artist); err != nil {
		return nil, err
	}
	return &artist, nil
}

func (m *MossDao) GetTrack(id string) (*Track, error) {
	var track Track
	if err := m.getRecord(fmt.Sprintf("track:by-id:%s", id), &track); err != nil {
		return nil, err
	}
	return &track, nil
}

func (m *MossDao) getRecords(prefix string, destination interface{}) error {
	if reflect.TypeOf(destination).Kind() != reflect.Ptr {
		return fmt.Errorf("getRecords called with non-pointer destination")
	}
	if reflect.Indirect(reflect.ValueOf(destination)).Kind() != reflect.Slice {
		return fmt.Errorf("getRecords called with non-slice pointer destination")
	}

	recordType := reflect.TypeOf(destination).Elem().Elem()
	valueOfDestination := reflect.ValueOf(destination)
	destinationElem := valueOfDestination.Elem()

	ss, err := m.collection.Snapshot()
	if err != nil {
		return err
	}
	defer ss.Close()

	iterator, err := ss.StartIterator([]byte(prefix), nil, moss.IteratorOptions{})
	if err != nil {
		return err
	}

	var iteratorError error
	for iteratorError == nil {
		var key, value []byte
		key, value, iteratorError = iterator.Current()

		if !strings.HasPrefix(string(key), prefix) {
			iteratorError = moss.ErrIteratorDone
			continue
		}

		var reader = bytes.NewBuffer(value)
		var record = reflect.New(recordType).Interface()
		if err := json.NewDecoder(reader).Decode(record); err != nil {
			return err
		}

		destinationElem.Set(
			reflect.Append(
				destinationElem,
				reflect.Indirect(reflect.ValueOf(record)).Convert(recordType),
			),
		)

		iteratorError = iterator.Next()
	}

	return nil
}

func (m *MossDao) getRecord(key string, destination interface{}) error {
	if reflect.TypeOf(destination).Kind() != reflect.Ptr {
		return fmt.Errorf("getRecord called with non-pointer destination")
	}

	ss, err := m.collection.Snapshot()
	if err != nil {
		return err
	}
	defer ss.Close()

	value, err := ss.Get([]byte(key), moss.ReadOptions{})
	if err != nil {
		log.WithError(err).Errorf("error getting record %q", key)
		return err
	}

	var reader = bytes.NewBuffer(value)
	if err := json.NewDecoder(reader).Decode(&destination); err != nil {
		log.WithError(err).Errorf("error decoding record %q [%s]", string(value), string(key))
		return err
	}

	return nil
}

func (m *MossDao) Close() error {
	if err := m.collection.Close(); err != nil {
		return err
	}

	return nil
}
