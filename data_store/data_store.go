package data_store

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
	log "github.com/sirupsen/logrus"
)

type Album struct {
	Id     string   `json:"id"`
	Name   string   `json:"name"`
	Tracks []*Track `json:"tracks"`
}

type Artist struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Track struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	AbsolutePath string `json:"absolutePath" api:"exclude"`
	TrackNumber  int    `json:"trackNumber"`
	Extension    string `json:"extension"`
}

type Dao interface {
	AddTrack(album *Album, artist *Artist, track *Track) error
	GetAlbums() ([]*Album, error)
	GetAlbum(id string) (*Album, error)
	GetArtists() ([]*Artist, error)
	GetArtist(id string) (*Artist, error)
	GetTrack(id string) (*Track, error)
	Close() error
}

// DataStore is a struct that represents the underlying database
type DataStore struct {
	sourceDirectory string
	dao             Dao
	searchClient    *SearchClient
	scanInProgress  bool
}

// NewDataStore returns a new data store
func NewDataStore(sourceDirectory, persistenceDirectory string, searchClient *SearchClient) *DataStore {
	return &DataStore{
		sourceDirectory: sourceDirectory,
		dao:             NewMossDao(persistenceDirectory),
		searchClient:    searchClient,
		scanInProgress:  false,
	}
}

// Scan analyzes the source directory to update the data store
func (d *DataStore) Scan() {
	if d.scanInProgress {
		return
	}
	d.scanInProgress = true
	d.searchClient.DropAll()
	d.scanDirectory(d.sourceDirectory)
}

func (d *DataStore) GetAlbums() ([]*Album, error) {
	return d.dao.GetAlbums()
}

func (d *DataStore) GetAlbum(id string) (*Album, error) {
	return d.dao.GetAlbum(id)
}

func (d *DataStore) GetArtists() ([]*Artist, error) {
	return d.dao.GetArtists()
}

func (d *DataStore) GetArtist(id string) (*Artist, error) {
	return d.dao.GetArtist(id)
}

func (d *DataStore) GetTrack(id string) (*Track, error) {
	return d.dao.GetTrack(id)
}

// scanDirectory is a helper function to recursively scan directories for data
func (d *DataStore) scanDirectory(root string) {
	files, err := ioutil.ReadDir(root)
	if err != nil {
		log.WithError(err).Fatalf("error scanning source directory %q", d.sourceDirectory)
	}

	for _, file := range files {
		absolutePath := filepath.Join(root, file.Name())
		if file.IsDir() {
			d.scanDirectory(absolutePath)
			continue
		}

		extension := strings.TrimLeft(strings.ToLower(filepath.Ext(file.Name())), ".")

		switch extension {
		case "mp3":
		case "m4a":
			// Great
		default:
			continue
		}

		fileHandler, err := os.Open(absolutePath)
		if err != nil {
			log.WithError(err).Warnf("error opening %q", absolutePath)
		}
		defer fileHandler.Close()

		m, err := tag.ReadFrom(fileHandler)
		if err != nil {
			log.WithError(err).Warnf("error reading tags for %q", absolutePath)
		} else {
			trackNumber, _ := m.Track()

			if m.Title() == "" {
				log.Warnf("no track information %q %q", m.Title(), absolutePath)
				continue
			}

			var album *Album
			var artist *Artist

			track := NewTrack(m.Album(), m.Artist(), m.Title(), trackNumber, absolutePath, extension)
			if m.Album() != "" {
				album = NewAlbum(m.Album(), m.Artist())
				if err := d.searchClient.AddDocument("album", map[string]interface{}{
					"Id":   album.Id,
					"Name": album.Name,
				}); err != nil {
					log.WithError(err).Errorf("error adding album to index")
				}
			}
			if m.Artist() != "" {
				artist = NewArtist(m.Artist(), m.Album())
				if err := d.searchClient.AddDocument("artist", map[string]interface{}{
					"Id":   artist.Id,
					"Name": artist.Name,
				}); err != nil {
					log.WithError(err).Errorf("error adding artist to index")
				}
			}

			if err := d.dao.AddTrack(album, artist, track); err != nil {
				log.WithError(err).Warnf("error adding song to data store %q", absolutePath)
			}
			if err := d.searchClient.AddDocument("track", map[string]interface{}{
				"Id":   track.Id,
				"Name": track.Name,
			}); err != nil {
				log.WithError(err).Errorf("error adding track to index")
			}
		}
	}

	if err := d.searchClient.FlushBatches(); err != nil {
		log.WithError(err).Error("error flushing search batches")
	}

	d.scanInProgress = false
}

func NewTrack(album, artist, track string, trackNumber int, absolutePath, extension string) *Track {
	return &Track{
		AbsolutePath: absolutePath,
		Extension:    extension,
		Id:           md5Hash(absolutePath),
		Name:         track,
		TrackNumber:  trackNumber,
	}
}

func NewAlbum(album, artist string) *Album {
	return &Album{
		Id:   md5Hash(fmt.Sprintf("%s%s", artist, album)),
		Name: album,
	}
}

func NewArtist(artist, album string) *Artist {
	return &Artist{
		Id:   md5Hash(fmt.Sprintf("%s%s", album, artist)),
		Name: artist,
	}
}
