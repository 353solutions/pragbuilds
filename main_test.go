package main

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_parseHTML(t *testing.T) {
	file, err := os.Open("testdata/builds.html")
	require.NoError(t, err, "open")
	defer file.Close()

	builds, err := parseHTML(file)
	require.NoError(t, err, "parse")
	require.Equal(t, 43, len(builds), "count")

	for i, b := range builds {
		require.NotEqualf(t, "", b.Name, "%d name", i)
		require.NotEqualf(t, 0, b.ID, "%d id", i)
	}
}

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
	notify(context.Background(), builds, observers, &s)
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
