package database

import (
	"log"

	"github.com/boltdb/bolt"
)

// Database represents a BoltDB database object.
//
// Contains the bucket name and the DB object for the BoltDB database.
type Database struct {
	Bucket string
	BoltDB *bolt.DB
}

// SetupDB opens a Bolt Database and creates a Bucket for storing
// key-value pairs.
//
// If the dbName file does not exist it will create it.
func SetupDB(dbName string, bucket string) (*Database, error) {
	// Open the db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		return nil, err
	}
	// Create the buckets, if they do not exist
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		log.Printf("Bolt Bucket \"%s\", setup done", bucket)
		return nil
	})
	if err != nil {
		return nil, err
	}
	log.Printf("Bolt Database \"%s\", setup done", dbName)
	return &Database{
		BoltDB: db,
		Bucket: bucket,
	}, nil
}

// PutEntryDB inserts a new key-value pair into the Bolt Database.
func PutEntryDB(db *Database, key string, value string) error {
	err := db.BoltDB.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte(db.Bucket)).Put([]byte(key), []byte(value))
		if err != nil {
			return err
		}
		log.Printf("Bolt Bucket \"%s\", updated with (\"%s\":\"%s\")", db.Bucket, key, value)
		return nil
	})
	if err != nil {
		return err
	}
	return err
}

// GetEntryDB reads a key-value pair from the Bolt Database Bucket,
// given the key.
func GetEntryDB(db *Database, key string) (string, error) {
	value := ""
	err := db.BoltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Bucket))
		v := b.Get([]byte(key))
		value = string(v)
		return nil
	})
	if err != nil {
		return "", err
	}
	return value, err
}

// GetEntriesDB reads all key-value pairs from the Bolt Database Bucket.
func GetEntriesDB(db *Database) (map[string]string, error) {
	entries := make(map[string]string)
	err := db.BoltDB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(db.Bucket))
		b.ForEach(func(k, v []byte) error {
			entries[string(k)] = string(v)
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entries, err
}

// DeleteEntryDB deletes a key-value pair from the Bolt Database Bucket,
// given the key.
func DeleteEntryDB(db *Database, key string) error {
	err := db.BoltDB.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte(db.Bucket)).Delete([]byte(key))
		if err != nil {
			return err
		}
		log.Printf("Bolt Bucket \"%s\", deleted \"%s\" key", db.Bucket, key)
		return nil
	})
	if err != nil {
		return err
	}
	return err
}

// PutMapEntriesDB inserts a map of key-value pairs into the Bolt Database.
func PutMapEntriesDB(db *Database, entries map[string]string) error {
	err := db.BoltDB.Update(func(tx *bolt.Tx) error {
		for key, value := range entries {
			err := tx.Bucket([]byte(db.Bucket)).Put([]byte(key), []byte(value))
			if err != nil {
				return err
			}
			log.Printf("Bolt Bucket \"%s\", updated with (\"%s\":\"%s\")", db.Bucket, key, value)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return err
}
