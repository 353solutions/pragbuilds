package main

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	dbFile := path.Join(t.TempDir(), "notifications.json")

	db, err := OpenDB(dbFile)
	require.NoError(t, err, "open")
	defer db.Close()

	id, email := 7, "a@example.com"
	db.Add(id, email)
	ok := db.Has(id, email)
	require.True(t, ok, "has")

	ok = db.Has(id+3, email)
	require.False(t, ok, "has not")

	db.Close()
	db, err = OpenDB(dbFile)
	require.NoError(t, err, "open 2")
	ok = db.Has(id, email)
	require.True(t, ok, "has 2")
}
