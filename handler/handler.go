package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
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

// parseJSON takes the Reponse body and a ShortenReq object
// and unmarshalls the JSON reponse to the ShortenReq object.
func parseJSON(reqBody io.Reader, req *ShortenReq) error {
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

// sendResponse replies to the request with a JSON response and a status code.
func sendResponse(w http.ResponseWriter, r *http.Request, statusCode int, jsonResp []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(jsonResp)
}

// sendErrorResponse same as sendResponse but creates and logs the error message to send.
func sendErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, e error) {
	log.Print(e)
	jsonResp, err := createJSON(ErrorResp{Error: e.Error()})
	if err != nil {
		log.Fatal(err)
	}
	sendResponse(w, r, statusCode, jsonResp)
}

// NewRouter creates a new httprouter Router,
// sets the handler for each route and adds middleware
// for enabling CORS with the default options.
//
// Returns an http.Handler
func NewRouter(db *database.Database) http.Handler {
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/redirect", Redirect)
	router.GET("/redirect/:short_url", RedirectURL(db))
	router.POST("/shorten", ShortenURL(db))
	handler := cors.Default().Handler(router)
	return handler
}

// Index handles requests for / path
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	jsonResp, err := createJSON(MessageResp{Message: "Welcome to url-shortener"})
	if err != nil {
		log.Fatal(err)
	}
	sendResponse(w, r, http.StatusOK, jsonResp)
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
			sendErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		// Long URL Not Found
		if longURL == "" {
			msg := fmt.Sprintf("URL %s Not Found", shortURL)
			jsonResp, err := createJSON(MessageResp{Message: msg})
			if err != nil {
				log.Fatal(err)
			}
			sendResponse(w, r, http.StatusNotFound, jsonResp)
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
		err := parseJSON(r.Body, &req)
		defer r.Body.Close()
		if err != nil {
			// Error Reading Request Body
			sendErrorResponse(w, r, http.StatusBadRequest, err)
			return
		}
		// Long URL Not Supplied
		if req.LongURL == "" {
			jsonResp, err := createJSON(ErrorResp{Error: "long_url parameter is required"})
			if err != nil {
				log.Fatal(err)
			}
			sendResponse(w, r, http.StatusBadRequest, jsonResp)
			return
		}
		// Convert long_url to short_url
		longURL, shortURL, err := shortener.Encode(req.LongURL)
		if err != nil {
			// Encoding Error
			sendErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}

		// Check if (short_url, long_url) entry is in Database (To detect Collision Error)
		longURLTest, err := database.GetEntryDB(db, shortURL)
		if err != nil {
			// Database Error
			sendErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		// If same short_url (key) is found but different long_url (value), Collision Error
		if longURLTest != longURL && longURLTest != "" {
			msg := fmt.Sprintf("URL Collision, Found 2 values for %s key: %s, %s", shortURL, longURLTest, longURL)
			log.Print(msg)
			// TMP: Send JSON response
			jsonResp, err := createJSON(MessageResp{Message: msg})
			if err != nil {
				log.Fatal(err)
			}
			sendResponse(w, r, http.StatusInternalServerError, jsonResp)
			return
		}

		// Save to Database
		err = database.PutEntryDB(db, shortURL, longURL)
		if err != nil {
			// Database Error
			sendErrorResponse(w, r, http.StatusInternalServerError, err)
			return
		}
		// Return short_url in JSON
		jsonResp, err := createJSON(ShortenResp{
			LongURL:  longURL,
			ShortURL: shortURL,
		})
		if err != nil {
			log.Fatal(err)
		}
		sendResponse(w, r, http.StatusOK, jsonResp)
	}
}
