package main

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strconv"

	"github.com/williamhaley/music-server/data_store"
	"github.com/williamhaley/music-server/server"
	log "github.com/sirupsen/logrus"
)

var accessToken = os.Getenv("ACCESS_TOKEN")
var musicDirectory = os.Getenv("MUSIC_DIRECTORY")
var persistenceDirectory = os.Getenv("PERSISTENCE_DIRECTORY")
var meilisearchAddress = os.Getenv("MEILISEARCH_ADDRESS")
var portString = os.Getenv("PORT")
var port int = 80

//go:embed web-client/build
var content embed.FS

// init parses and evaluates command flags
func init() {
	if len(accessToken) < 8 {
		fmt.Println("ACCESS_TOKEN must be at least 8 characters")
		os.Exit(1)
	}
	if info, err := os.Stat(musicDirectory); err != nil || !info.IsDir() {
		fmt.Println("MUSIC_DIRECTORY must be a valid directory")
		os.Exit(1)
	}
	if info, err := os.Stat(persistenceDirectory); err != nil || !info.IsDir() {
		fmt.Println("PERSISTENCE_DIRECTORY must be a valid directory")
		os.Exit(1)
	}
	if meilisearchAddress == "" {
		fmt.Println("MEILISEARCH_ADDRESS must be a valid string")
		os.Exit(1)
	}
	if portString != "" {
		if parsedPort, err := strconv.Atoi(portString); err != nil {
			fmt.Println("PORT must be a valid port number")
			os.Exit(1)
		} else {
			port = parsedPort
		}
	}
}

// main runs the application
func main() {
	searchClient := data_store.NewSearchClient(meilisearchAddress)
	dataStore := data_store.NewDataStore(musicDirectory, persistenceDirectory, searchClient)

	log.Info("starting server")

	staticFiles, err := fs.Sub(content, "web-client/build")
	if err != nil {
		panic(err)
	}
	server := server.NewServer(dataStore, searchClient, accessToken, http.FS(staticFiles))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), server); err != nil {
		log.WithError(err).Fatal("error running server")
	}
}
