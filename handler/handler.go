package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

// ShortenResp represents the JSON response that will be sent,
// on the shorten endpoint.
type ShortenResp struct {
	LongURL  string `json:"long_url"`
	ShortURL string `json:"short_url"`
}

// MessageResp represents the JSON response that will be sent as a message.
type MessageResp struct {
	Message string `json:"message"`
}

// ErrorResp represents the JSON response that will be sent on errors.
type ErrorResp struct {
	Error string `json:"error"`
}

// createJSON takes any Resp type as an argument
// and returns a JSON response with its fields as attributes and values.
func createJSON(r interface{}) ([]byte, error) {
	jsonResp, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return jsonResp, nil
}

// parseJSONRequest takes the Reponse body and a ShortenReq object
// and unmarshalls the JSON reponse to the ShortenReq object.
func parseJSONRequest(reqBody io.Reader, req *ShortenReq) error {
	body, err := ioutil.ReadAll(reqBody)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, req)
	if err != nil {
		return err
	}
	return nil
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
	msgResp := MessageResp{Message: "Welcome to url-shortener"}
	jsonResp, err := createJSON(msgResp)
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
			errorResp := ErrorResp{Error: err.Error()}
			jsonResp, err := createJSON(errorResp)
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
			msg := fmt.Sprintf("URL %s Not Found", shortURL)
			msgResp := MessageResp{Message: msg}
			jsonResp, err := createJSON(msgResp)
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
		// Read Request Body, parse it as JSON and get long URL
		err := parseJSONRequest(r.Body, &req)
		defer r.Body.Close()
		if err != nil {
			// Error Reading Request Body
			log.Print(err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			errorResp := ErrorResp{Error: err.Error()}
			jsonResp, err := createJSON(errorResp)
			if err != nil {
				log.Fatal(err)
			}
			w.Write(jsonResp)
			return
		}

		// Long URL Not Supplied
		if req.LongURL == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			errorResp := ErrorResp{Error: "long_url parameter is required"}
			jsonResp, err := createJSON(errorResp)
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
			errorResp := ErrorResp{Error: err.Error()}
			jsonResp, err := createJSON(errorResp)
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
			errorResp := ErrorResp{Error: err.Error()}
			jsonResp, err := createJSON(errorResp)
			if err != nil {
				log.Fatal(err)
			}
			w.Write(jsonResp)
			return
		}

		// Return short_url in JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		shortenResp := ShortenResp{LongURL: req.LongURL, ShortURL: shortURL}
		jsonResp, err := createJSON(shortenResp)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(jsonResp)
	}
}
