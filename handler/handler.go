package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/thanoskoutr/url-shortener/database"
)

// createJSONResponse takes string slices of attributes and its values
// and returns a JSON reponse with the message, as a slice of bytes.
func createJSONResponse(attributes []string, values []string) ([]byte, error) {
	resp := make(map[string]string)
	for i, attribute := range attributes {
		resp[attribute] = values[i]
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return jsonResp, nil
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

// Index handles requests for / path
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	w.Header().Set("Content-Type", "application/json")
	attr := []string{"message"}
	val := []string{"Welcome to url-shortener"}
	jsonResp, err := createJSONResponse(attr, val)
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
		shortURL := ps.ByName("short_url")

		longURL, err := database.GetEntryDB(db, shortURL)
		if err != nil {
			// Database Error
			log.Fatal(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			attr := []string{"error"}
			val := []string{err.Error()}
			jsonResp, err := createJSONResponse(attr, val)
			if err != nil {
				log.Fatal(err)
			}
			w.Write(jsonResp)
		}

		if longURL == "" {
			// URL Not Found
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			attr := []string{"message"}
			msg := fmt.Sprintf("URL %s Not Found", shortURL)
			val := []string{msg}
			jsonResp, err := createJSONResponse(attr, val)
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
		longURL := ps.ByName("long_url")
		// Check if in Database -> Return short_url, Else continue
		// Convert long_url to short_url -> Return short_url in JSON

		// TEMP: Just returns the request
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		attr := []string{"long_url", "short_url"}
		val := []string{longURL, "lol"}
		jsonResp, err := createJSONResponse(attr, val)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(jsonResp)
	}
}
