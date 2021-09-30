package handler

import (
	"net/http"

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
			http.Redirect(w, r, url, http.StatusFound)

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
