package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/thanoskoutr/url-shortener/database"
	"github.com/thanoskoutr/url-shortener/handler"
)

const (
	// dbFlagName represents the name of the database flag
	dbFlagName = "db"
	// dbFlagValue represents the default value of the database flag
	dbFlagValue = "urls.db"
	// dbFlagUsage represents the usage string of the database flag
	dbFlagUsage = "Database file"
	// logFlagName represents the name of the log flag
	logFlagName = "log"
	// logFlagValue represents the default value of the log flag
	logFlagValue = "logs.txt"
	// logFlagUsage represents the usage string of the log flag
	logFlagUsage = "Log file"

	// BUCKET_NAME represents the name of the Bucket in the Bolt Database
	BUCKET_NAME = "URL"

	// PORT represents the default port the server will listen
	PORT = "8080"
)

func main() {
	// Parse command-line flag
	// yamlFilename := flag.String("yaml", "urls.yaml", "YAML file with URLs and their short paths")
	// jsonFilename := flag.String("json", "urls.json", "JSON file with URLs and their short paths")
	dbFilename := flag.String(dbFlagName, dbFlagValue, dbFlagUsage)
	logFilename := flag.String(logFlagName, logFlagValue, logFlagUsage)
	flag.Parse()

	// Seutp Logging
	// If the file doesn't exist, create it or append to the file
	file, err := os.OpenFile(*logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	defer file.Close()
	log.Printf("Logging setup done\n")

	// Setup Database
	db, err := database.SetupDB(*dbFilename, BUCKET_NAME)
	if err != nil {
		log.Fatal(err)
	}
	defer db.BoltDB.Close()

	// Create a default request multiplexer as the last fallback
	mux := defaultMux()

	// Build the DBHandler using the previous handler as the fallback
	dbHandler := createDBHandler(db, mux)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = PORT
	}
	fmt.Printf("Starting the server on :%v", port)
	log.Printf("Listening on port: %v", port)
	if err := http.ListenAndServe(":"+port, dbHandler); err != nil {
		log.Fatal(err)
	}
}

// defaultMux is a default request multiplexer for all paths
func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	return mux
}

// home is the default fallaback function handler for all paths
func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	resp := make(map[string]string)
	resp["message"] = "Resource Not Found"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(jsonResp)
}

// createDBHandler reads the Database, creates and returns a DB Hundler
func createDBHandler(db *database.Database, fallback http.Handler) http.HandlerFunc {
	// Add some entries to the Database
	pathsToUrls := map[string]string{
		"/gnu/health":   "https://savannah.gnu.org/projects/health",
		"/gnu/avr-libc": "https://savannah.nongnu.org/projects/avr-libc",
		"/gnu/dino":     "https://savannah.nongnu.org/projects/dino",
		"/gnu/ddd":      "https://savannah.gnu.org/projects/ddd",
		"/gnu/epsilon":  "https://savannah.gnu.org/projects/epsilon",
	}
	err := database.PutMapEntriesDB(db, pathsToUrls)
	if err != nil {
		log.Fatal(err)
	}

	dbHandler, err := handler.DBHandler(db, fallback)
	if err != nil {
		log.Fatal(err)
	}
	return dbHandler
}
