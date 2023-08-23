package main

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type Email struct {
	to      string
	content string
}

type MockSender struct {
	emails []Email
}

func (m *MockSender) Send(ctx context.Context, to, content string) error {
	e := Email{
		to:      to,
		content: content,
	}
	m.emails = append(m.emails, e)
	return nil
}

func Test_nofity(t *testing.T) {
	observers := map[string][]string{
		"Book 1": {"a@example.com", "b@example.com"},
		"Book 2": {"c@example.com"},
	}
	builds := []Build{
		{1, "Book 1", true},
		{2, "Book 1", false},
		{3, "Book 2", true},
	}

	var s MockSender
	db, err := OpenDB("/dev/null")
	require.NoError(t, err, "db")
	notify(context.Background(), builds, observers, db, &s)
	require.Equal(t, 3, len(s.emails), "email count")
}

func Test_fetchBuilds(t *testing.T) {
	user, passwd := os.Getenv("PRAG_USER"), os.Getenv("PRAG_PASSWD")
	if user == "" || passwd == "" {
		t.Skip("missing credentials")
	}

	var ctx context.Context
	var cancel context.CancelFunc
	d, ok := t.Deadline()
	if ok {
		ctx, cancel = context.WithDeadline(context.Background(), d)
	} else {
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	}
	defer cancel()

	r, err := fetchBuilds(ctx, user, passwd)
	require.NoError(t, err, "fetch")
	defer r.Close()

	data, err := io.ReadAll(io.LimitReader(r, 1_000_000))
	require.NoError(t, err, "read")
	html := string(data)
	require.Contains(t, html, `<table id="bookshelf"`, "html")
}
