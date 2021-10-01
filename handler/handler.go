package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/thanoskoutr/url-shortener/database"
	"github.com/thanoskoutr/url-shortener/shortener"
)

// ShortenReq represents the JSON body of the request that will be decoded,
// on the shorten endpoint.
type ShortenReq struct {
	LongURL string `json:"long_url"`
}

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
	router.POST("/shorten", ShortenURL(db))
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

// RedirectURL handles requests for /redirect/:short_url path
func RedirectURL(db *database.Database) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Short URL named parameter
		shortURL := ps.ByName("short_url")
		// Search Database
		longURL, err := database.GetEntryDB(db, shortURL)
		if err != nil {
			// Database Error
			log.Print(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			attr := []string{"error"}
			val := []string{err.Error()}
			jsonResp, err := createJSONResponse(attr, val)
			if err != nil {
				log.Fatal(err)
			}
			w.Write(jsonResp)
			return
		}
		// Long URL Not Found
		if longURL == "" {
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

// ShortenURL handles requests for /shorten path
func ShortenURL(db *database.Database) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Long URL named parameter
		var req ShortenReq
		// Read Body
		err := json.NewDecoder(r.Body).Decode(&req)
		defer r.Body.Close()
		if err != nil {
			// Error Parsing JSON body
			log.Print(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			attr := []string{"error"}
			val := []string{err.Error()}
			jsonResp, err := createJSONResponse(attr, val)
			if err != nil {
				log.Fatal(err)
			}
			w.Write(jsonResp)
			return
		}

		log.Printf("long_url = %s", req.LongURL)

		// Long URL Not Supplied
		if req.LongURL == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			attr := []string{"error"}
			val := []string{"long_url parameter is required"}
			jsonResp, err := createJSONResponse(attr, val)
			if err != nil {
				log.Fatal(err)
			}
			w.Write(jsonResp)
			return
		}
		// Convert long_url to short_url
		shortURL, err := shortener.Encode(req.LongURL)
		if err != nil {
			// Encoding Error
			log.Print(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			attr := []string{"error"}
			val := []string{err.Error()}
			jsonResp, err := createJSONResponse(attr, val)
			if err != nil {
				log.Fatal(err)
			}
			w.Write(jsonResp)
			return
		}

		// Save to Database
		err = database.PutEntryDB(db, shortURL, req.LongURL)
		if err != nil {
			// Database Error
			log.Print(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			attr := []string{"error"}
			val := []string{err.Error()}
			jsonResp, err := createJSONResponse(attr, val)
			if err != nil {
				log.Fatal(err)
			}
			w.Write(jsonResp)
			return
		}

		// Return short_url in JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		attr := []string{"long_url", "short_url"}
		val := []string{req.LongURL, shortURL}
		jsonResp, err := createJSONResponse(attr, val)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(jsonResp)
	}
}
