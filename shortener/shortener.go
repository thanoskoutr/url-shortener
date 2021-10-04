package shortener

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"log"
	"net/url"
)

// Encode takes a URL and converts it in a 7 character string
// using a hash function and base64 conversion.
//
// Returns the short URL and the long URL (absolute)
func Encode(longURL string) (string, string, error) {
	// Parse the URL
	u, err := url.Parse(longURL)
	if err != nil {
		return "", "", err
	}
	// Check if the URL is not absolute - has not a valid scheme (http or https)
	if !u.IsAbs() {
		// Add a scheme
		u.Scheme = "https"
	}
	// Convert URL to string
	longURLAbs := u.String()
	log.Printf("Shortener, long_url = %s", longURLAbs)

	// MD5 Hash of URL
	longURLHash := md5.New()
	io.WriteString(longURLHash, longURLAbs)
	// Save hash to byte slice
	longURLHashBytes := longURLHash.Sum(nil)
	log.Printf("Shortener, long_url (Hash) = %x\n", longURLHashBytes)

	// Base64 (URL Safe) encode it
	longURLHashBytesB64Safe := base64.RawURLEncoding.EncodeToString(longURLHashBytes)
	log.Printf("Shortener, long_url (Hash+Base64) = %s", longURLHashBytesB64Safe)

	// Keep the first 7 characters
	shortURLB64Safe := longURLHashBytesB64Safe[:7]
	log.Printf("Shortener, short_url = %s", shortURLB64Safe)

	// Return short URL
	return longURLAbs, shortURLB64Safe, nil
}

// Decode takes a short URL (7 character base64 hash) and converts back
// to hash in bytes.
func Decode(shortURL string) ([]byte, error) {
	// Base64 decode the 7 char string
	shortURLHashBytes, err := base64.RawURLEncoding.DecodeString(shortURL)
	if err != nil {
		return nil, err
	}
	return shortURLHashBytes, nil
}
