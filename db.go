package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"time"
)

type DBKey struct {
	BuildID int    `json:"build_id"`
	Email   string `json:"email"`
}

type DBEntry struct {
	DBKey
	Time time.Time `json:"time"`
}

type DB struct {
	fileName string
	entries  map[DBKey]time.Time
}

func OpenDB(fileName string) (*DB, error) {
	db := DB{
		fileName: fileName,
		entries:  make(map[DBKey]time.Time),
	}
	file, err := os.Open(fileName)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &db, nil
		}
		return nil, err
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	lnum := 0
	for {
		lnum++
		var e DBEntry
		err := dec.Decode(&e)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		db.entries[e.DBKey] = e.Time
	}

	return &db, nil
}

func (db *DB) Close() error {
	file, err := os.Create(db.fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	for key, time := range db.entries {
		e := DBEntry{
			DBKey: key,
			Time:  time,
		}
		if err := enc.Encode(e); err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) Add(buildID int, email string) {
	key := DBKey{buildID, email}
	db.entries[key] = time.Now().UTC()
}

func (db *DB) Has(buildID int, email string) bool {
	_, ok := db.entries[DBKey{buildID, email}]
	return ok
}
