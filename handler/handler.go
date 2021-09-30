package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/thanoskoutr/url-shortener/database"
)

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortPath := r.URL.Path
		if url, ok := pathsToUrls[shortPath]; ok {
			http.Redirect(w, r, url, http.StatusMovedPermanently)

		} else {
			fallback.ServeHTTP(w, r)
		}
	}
}

// DBHandler will query the Database and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the Database, then the
// fallback http.Handler will be called instead.
//
// Database is expected to be in key-value pair format.
//
// The only errors that can be returned all related to getting
// error from the Database.
func DBHandler(db *database.Database, fallback http.Handler) (http.HandlerFunc, error) {
	pathMap := make(map[string]string)
	// Read all entries from Database, save in a map
	entries, err := database.GetEntriesDB(db)
	if err != nil {
		return nil, err
	}
	for k, v := range entries {
		pathMap[k] = v
	}
	return MapHandler(pathMap, fallback), nil
}

// NewRouter creates a new httprouter Router and
// sets the handler for each route.
//
// Returns an *httprouter.Router
func NewRouter(db *database.Database) *httprouter.Router {
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/redirect", Redirect)
	router.GET("/redirect/:short_url", RedirectURL(db))
	router.GET("/shorten", Shorten)
	router.GET("/shorten/:long_url", ShortenURL(db))
	return router
}

// createJSONResponse takes a string message and returns a JSON reponse
// with the message, as a slice of bytes.
func createJSONResponse(attribute string, value string) ([]byte, error) {
	resp := make(map[string]string)
	resp[attribute] = value
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return jsonResp, nil
}

// Index handles requests for / path
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	msg := "Hello"
	jsonResp, err := createJSONResponse("message", msg)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(jsonResp)
}

// Redirect handles requests for /redirect path
func Redirect(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

// RedirectURL handles requests for /redirect/:url path
func RedirectURL(db *database.Database) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// TEMP: Dummy way, check directly in Database, do not read all entries and then check
		// FIX: GetEntryDB(shortURL) -> If found -> redirect to long URL, Else -> redirect home
		// Read all entries from Database, save in a map
		entries, err := database.GetEntriesDB(db)
		if err != nil {
			// Database Error
			log.Fatal(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			jsonResp, err := createJSONResponse("error", err.Error())
			if err != nil {
				log.Fatal(err)
			}
			w.Write(jsonResp)
		}
		log.Print(entries)

		shortURL := ps.ByName("short_url")
		longURL := entries[shortURL]
		// longURL := "https://savannah.nongnu.org/projects/dino"

		if longURL == "" {
			// URL Not Found
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			msg := fmt.Sprintf("URL %s Not Found", shortURL)
			jsonResp, err := createJSONResponse("message", msg)
			if err != nil {
				log.Fatal(err)
			}
			w.Write(jsonResp)
			return
		}
		// Redirect using long URL found
		http.Redirect(w, r, longURL, http.StatusMovedPermanently)
	}
}

// Shorten handles requests for /encode path
func Shorten(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

// ShortenURL handles requests for /encode/:url path
func ShortenURL(db *database.Database) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		longURL := ps.ByName("long_url")
		msg := fmt.Sprintf("Cannot encode URL: %s", longURL)
		jsonResp, err := createJSONResponse("message", msg)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(jsonResp)
	}
}
