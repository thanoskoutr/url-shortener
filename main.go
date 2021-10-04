package main

import (
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

	// Create a new httprouter to handle routes
	router := handler.NewRouter(db)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = PORT
	}
	fmt.Printf("Starting the server on :%v\n", port)
	log.Printf("Listening on port: %v", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
