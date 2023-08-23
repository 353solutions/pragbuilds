package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Printf("INFO: getting builds")
	prag_user, prag_passwd := os.Getenv("PRAG_USER"), os.Getenv("PRAG_PASSWD")
	if prag_user == "" || prag_passwd == "" {
		log.Fatalf("error: no pragmatic credentials")
	}

	gmail_user, gmail_passwd := os.Getenv("GMAIL_USER"), os.Getenv("GMAIL_PASSWD")
	if gmail_user == "" || gmail_passwd == "" {
		log.Fatalf("error: no gmail credentials")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rc, err := fetchBuilds(ctx, prag_user, prag_passwd)
	if err != nil {
		log.Fatalf("error: fetch builds - %s", err)
	}
	defer rc.Close()

	builds, err := parseHTML(rc)
	if err != nil {
		log.Fatalf("error: can't parse HTML - %s", err)
	}
	log.Printf("info: %d builds", len(builds))

	s := GmailSender{gmail_user, gmail_passwd}
	if err := notify(context.Background(), builds, db, s); err != nil {
		log.Fatalf("error: can't notify - %s", err)
	}
}

const buildsURL = "https://portal.pragprog.com/build_statuses"

func fetchBuilds(ctx context.Context, user, passwd string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, buildsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("can't build request for %q - %w", buildsURL, err)
	}
	req.SetBasicAuth(user, passwd)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("can't GET %q - %w", buildsURL, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%q: bad status - %s", buildsURL, resp.Status)
	}

	return resp.Body, nil
}

func notify(ctx context.Context, builds []Build, observers map[string][]string, sender Sender) error {
	var errs []error
	for _, b := range builds {
		if !b.Failed {
			continue
		}

		for _, email := range observers[b.Name] {
			// FIXME: Don't double notify
			log.Printf("INFO: Notifying %q on %+v", email, b)
			content := formatEmail(email, b)
			if err := sender.Send(ctx, email, content); err != nil {
				log.Printf("ERROR: can't email %q - %s", email, err)
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

type Sender interface {
	Send(ctx context.Context, to, content string) error
}

type GmailSender struct {
	user   string
	passwd string
}

func (s GmailSender) Send(ctx context.Context, to, content string) error {
	auth := LoginAuth(s.user, s.passwd)
	// FIXME
	return smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		s.user,
		[]string{to},
		[]byte(content),
	)
}

var emailTemplate = `From: PragBuilds Bot <pragbuilds@gmail.com>
To: %s
Subject: [pragbuilds bot] %q build %d failed

Hi There,

I'm sorry to inform you that build %d of %q has failed.
You can see the logs at https://portal.pragprog.com/build_statuses/%d/log.

Onward and upward,
The pragbuilds bot
`

func formatEmail(to string, build Build) string {
	text := fmt.Sprintf(
		emailTemplate,
		to,
		build.Name, build.ID, // subject
		build.ID, build.Name, build.ID, // body
	)
	return text
}

// https://gist.github.com/jpillora/cb46d183eca0710d909a?permalink_comment_id=3519541
type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, errors.New("Unkown fromServer")
		}
	}
	return nil, nil
}

// Poor man's database
var db = map[string][]string{
	"Effective Go Recipes": {"miki@353solutions.com"},
}
