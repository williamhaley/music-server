package server

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	pa "github.com/williamhaley/go/api"
	"github.com/williamhaley/music-server/data_store"
	log "github.com/sirupsen/logrus"
)

// APIHandler is a struct that maintains state for API HTTP handler funcs
type APIHandler struct {
	dataStore    *data_store.DataStore
	searchClient *data_store.SearchClient
}

// NewAPIHandler returns a new instance of an APIHandler
func NewAPIHandler(dataStore *data_store.DataStore, searchClient *data_store.SearchClient) *APIHandler {
	return &APIHandler{
		dataStore:    dataStore,
		searchClient: searchClient,
	}
}

func (h *APIHandler) CheckAuth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *APIHandler) Search(w http.ResponseWriter, r *http.Request) {
	searchTerm := r.URL.Query().Get("q")

	results, _ := h.searchClient.Search(searchTerm)

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data": results,
	}); err != nil {
		http.Error(w, "error returning search results", http.StatusInternalServerError)
		log.WithError(err).Errorf("error returning search results %q", searchTerm)
		return
	}
}

func (h *APIHandler) Scan(w http.ResponseWriter, r *http.Request) {
	go h.dataStore.Scan()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("starting scan..."))
}

// GetAlbums returns information on albums
func (h *APIHandler) GetAlbums(w http.ResponseWriter, r *http.Request) {
	albums, err := h.dataStore.GetAlbums()
	if err != nil {
		http.Error(w, "error getting albums", http.StatusInternalServerError)
		log.WithError(err).Error("error getting albums")
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data": pa.ToApiData(albums),
	}); err != nil {
		http.Error(w, "error returning albums", http.StatusInternalServerError)
		log.WithError(err).Error("error returning albums")
		return
	}
}

// GetAlbum returns information on an album
func (h *APIHandler) GetAlbum(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	album, err := h.dataStore.GetAlbum(id)
	if err != nil {
		http.Error(w, "error getting album", http.StatusInternalServerError)
		log.WithError(err).Errorf("error getting album %q", id)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data": pa.ToApiData(album),
	}); err != nil {
		http.Error(w, "error returning album", http.StatusInternalServerError)
		log.WithError(err).Errorf("error returning album %q", id)
		return
	}
}

// GetArtists returns information on artists
func (h *APIHandler) GetArtists(w http.ResponseWriter, r *http.Request) {
	artists, err := h.dataStore.GetArtists()
	if err != nil {
		http.Error(w, "error getting artists", http.StatusInternalServerError)
		log.WithError(err).Error("error getting artists")
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data": pa.ToApiData(artists),
	}); err != nil {
		http.Error(w, "error returning artists", http.StatusInternalServerError)
		log.WithError(err).Error("error returning artists")
		return
	}
}

// GetArtist returns information on an album
func (h *APIHandler) GetArtist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	artist, err := h.dataStore.GetArtist(id)
	if err != nil {
		http.Error(w, "error getting artist", http.StatusInternalServerError)
		log.WithError(err).Errorf("error getting artist %q", id)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"data": pa.ToApiData(artist),
	}); err != nil {
		http.Error(w, "error returning artist", http.StatusInternalServerError)
		log.WithError(err).Errorf("error returning artist %q", id)
		return
	}
}

// GetTrack returns a track
func (h *APIHandler) GetTrack(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	track, err := h.dataStore.GetTrack(id)
	if err != nil {
		http.Error(w, "error getting track", http.StatusInternalServerError)
		log.WithError(err).Errorf("error getting track %q", id)
		return
	}

	file, err := os.Open(track.AbsolutePath)
	if err != nil {
		http.Error(w, "error loading track", http.StatusInternalServerError)
		log.WithError(err).Errorf("error loading track %q", id)
		return
	}
	defer file.Close()

	// Be cautious about changing this. Safari seems to be picky about range support for audio files.
	http.ServeFile(w, r, file.Name())
}
