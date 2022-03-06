package data_store

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/dhowden/tag"
	log "github.com/sirupsen/logrus"
	ds "github.com/williamhaley/go/data_store"
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
	mossDao         *ds.MossDao
	searchClient    *SearchClient
	scanInProgress  bool
}

// NewDataStore returns a new data store
func NewDataStore(sourceDirectory, persistenceDirectory string, searchClient *SearchClient) *DataStore {
	mossDao, err := ds.NewMossDao(persistenceDirectory)
	if err != nil {
		log.WithError(err).Fatal("error creating data store")
	}

	return &DataStore{
		sourceDirectory: sourceDirectory,
		mossDao:         mossDao,
		searchClient:    searchClient,
		scanInProgress:  false,
	}
}

// Scan analyzes the source directory to update the data store
func (d *DataStore) Scan() string {
	if d.scanInProgress {
		return "already in progress"
	}

	d.scanInProgress = true

	d.searchClient.DropAll()

	config := &ds.ScanConfiguration{}
	config.AddExtensions("mp3", "m4a")

	ds.ScanDirectory(d.sourceDirectory, config, func(result *ds.ScanResult) {
		fileHandler, err := os.Open(result.AbsolutePath)
		if err != nil {
			log.WithError(err).Warnf("error opening %q", result.AbsolutePath)
		}
		defer fileHandler.Close()

		m, err := tag.ReadFrom(fileHandler)
		if err != nil {
			log.WithError(err).Warnf("error reading tags for %q", result.AbsolutePath)
		} else {
			trackNumber, _ := m.Track()

			if m.Title() == "" {
				log.Warnf("no track information %q %q", m.Title(), result.AbsolutePath)
				return
			}

			var album *Album
			var artist *Artist

			track := NewTrack(m.Album(), m.Artist(), m.Title(), trackNumber, result.AbsolutePath, result.Extension)
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

			if err := d.AddTrack(album, artist, track); err != nil {
				log.WithError(err).Warnf("error adding song to data store %q", result.AbsolutePath)
			}
			if err := d.searchClient.AddDocument("track", map[string]interface{}{
				"Id":   track.Id,
				"Name": track.Name,
			}); err != nil {
				log.WithError(err).Errorf("error adding track to index")
			}
		}
	}, func(count int, err error) {
		d.scanInProgress = false

		if err != nil {
			log.WithError(err).Error("error scanning")
		} else {
			log.Infof("scan completed. found %d files", count)
		}
	})

	return "starting scan..."
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

func md5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func (d *DataStore) AddTrack(album *Album, artist *Artist, track *Track) error {
	if err := d.mossDao.SetRecord(fmt.Sprintf("track:by-id:%s", track.Id), track); err != nil {
		return err
	}

	if album != nil {
		trackSortKey := fmt.Sprintf("%06d:%s", track.TrackNumber, track.Id)
		if err := d.mossDao.SetRecord(fmt.Sprintf("track:for-album-id:%s:%s", album.Id, trackSortKey), track); err != nil {
			return err
		}
		if err := d.mossDao.SetRecord(fmt.Sprintf("album:by-name:%s", strings.ToLower(album.Name)), album); err != nil {
			return err
		}
		if err := d.mossDao.SetRecord(fmt.Sprintf("album:by-id:%s", album.Id), album); err != nil {
			return err
		}
	}

	if artist != nil {
		if album != nil {
			if err := d.mossDao.SetRecord(fmt.Sprintf("album:for-artist-id:%s:%s", artist.Id, album.Id), album); err != nil {
				return err
			}
		}
		if err := d.mossDao.SetRecord(fmt.Sprintf("artist:by-name:%s", strings.ToLower(artist.Name)), artist); err != nil {
			return err
		}
		if err := d.mossDao.SetRecord(fmt.Sprintf("artist:by-id:%s", artist.Id), artist); err != nil {
			return err
		}
	}

	return nil
}

func (d *DataStore) GetAlbums() ([]*Album, error) {
	var albums []*Album
	if err := d.mossDao.GetRecords("album:by-name:", &albums); err != nil {
		return nil, err
	}
	return albums, nil
}

func (d *DataStore) GetAlbum(id string) (*Album, error) {
	var album Album
	if err := d.mossDao.GetRecord(fmt.Sprintf("album:by-id:%s", id), &album); err != nil {
		return nil, err
	}
	var tracks []*Track
	if err := d.mossDao.GetRecords(fmt.Sprintf("track:for-album-id:%s:", album.Id), &tracks); err != nil {
		return nil, err
	}
	album.Tracks = tracks
	return &album, nil
}

func (d *DataStore) GetArtists() ([]*Artist, error) {
	var artists []*Artist
	if err := d.mossDao.GetRecords("artist:by-name:", &artists); err != nil {
		return nil, err
	}
	return artists, nil
}

func (d *DataStore) GetArtist(id string) (*Artist, error) {
	var artist Artist
	if err := d.mossDao.GetRecord(fmt.Sprintf("artist:by-id:%s", id), &artist); err != nil {
		return nil, err
	}
	return &artist, nil
}

func (d *DataStore) GetTrack(id string) (*Track, error) {
	var track Track
	if err := d.mossDao.GetRecord(fmt.Sprintf("track:by-id:%s", id), &track); err != nil {
		return nil, err
	}
	return &track, nil
}
