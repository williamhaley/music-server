package server

import (
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	pm "github.com/williamhaley/go/middleware"
	"github.com/williamhaley/music-server/data_store"
)

// NewServer returns an http.Handler to handle all web/HTTP API calls
func NewServer(dataStore *data_store.DataStore, searchClient *data_store.SearchClient, accessToken string, staticContentFilesystem http.FileSystem) http.Handler {
	server := chi.NewRouter()

	server.Use(middleware.Logger)

	apiHandler := NewAPIHandler(dataStore, searchClient)

	server.Group(func(r chi.Router) {
		r.Use(pm.NewAuthMiddleware(accessToken))

		r.Get("/api/auth", apiHandler.CheckAuth)

		r.Get("/api/search", apiHandler.Search)

		r.Get("/admin/api/scan", apiHandler.Scan)

		r.Get("/api/albums/{id}", apiHandler.GetAlbum)
		r.Get("/api/albums", apiHandler.GetAlbums)
		r.Get("/api/artists/{id}", apiHandler.GetArtist)
		r.Get("/api/artists", apiHandler.GetArtists)
		r.Get("/api/track/{id}", apiHandler.GetTrack)
	})

	webHandler := http.FileServer(staticContentFilesystem).ServeHTTP
	uiServerAddress := os.Getenv("UI_SERVER_ADDRESS")
	if uiServerAddress != "" {
		webHandler = uiServerProxy(uiServerAddress)
	}
	server.Get("/*", webHandler)

	return server
}
