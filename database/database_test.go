package database

import (
	"os"
	"testing"
)

// Testcases Maphandler
var pathsToUrls = map[string]string{
	"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
	"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
}

// Global database object pointer
var db *Database

func setup() {
	// Setup Database
	dbFilename := "urls.db"
	BUCKET_NAME := "URL"
	var err error
	db, err = SetupDB(dbFilename, BUCKET_NAME)
	if err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	setup()
	os.Exit(m.Run())
}

func TestPutEntryDB(t *testing.T) {
	// Put entry, key-value pair
	k := "/ghb/urlshort"
	v := "https://github.com/gophercises/urlshort"
	err := PutEntryDB(db, k, v)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Entry added, key: %s, value: %s\n", k, v)

	k = "/ghb/authelia"
	v = "https://github.com/authelia/authelia"
	err = PutEntryDB(db, k, v)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Entry added, key: %s, value: %s\n", k, v)
}

func TestGetEntryDB(t *testing.T) {
	// Get entry by key
	k := "/ghb/urlshort"
	v, err := GetEntryDB(db, k)
	if err != nil {
		t.Fatal(err)
	}
	if v == "" {
		t.Errorf("no value for: %s\n", k)
	}
	t.Logf("key: %s, value: %s\n", k, v)
}

func TestGetEntriesDB(t *testing.T) {
	// Get all entries
	entries, err := GetEntriesDB(db)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range entries {
		t.Logf("key: %s, value: %s\n", k, v)
	}
}

func TestDeleteEntryDB(t *testing.T) {
	// Delete entry
	k := "/ghb/urlshort"
	err := DeleteEntryDB(db, k)
	if err != nil {
		t.Fatal(err)
	}
	// Check if still in database
	v, err := GetEntryDB(db, k)
	if err != nil {
		t.Fatal(err)
	}
	if v != "" {
		t.Errorf("Entry still in database, key: %s\n", k)
	}
	t.Logf("Entry deleted, key: %s\n", k)
}

func TestPutMapEntriesDB(t *testing.T) {
	// Add a map of URLs to Database
	err := PutMapEntriesDB(db, pathsToUrls)
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range pathsToUrls {
		t.Logf("Entry added, key: %s, value: %s\n", k, v)
	}
}
