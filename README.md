# PragBuilds Bot

A bot that monitors [https://portal.pragprog.com/build_statuses][ps] and report failed builds to authors.
It runs on cron as a GitHub action (see [here](.github/workflows/cron.yml)).

## Credentials

You need to set the following secrets on GitHub actions or environment variables for local development.

```
export PRAG_USER=XXX
export PRAG_PASSWD=XXX
export GMAIL_USER=XXX
export GMAIL_PASSWD=XXX
```

[ps]: https://portal.pragprog.com/build_statuses

## Get Notified

To add another book, edit `db` in `main.go`.

## Debugging

When debugging, set `export DEBUG=yes` environment variable to have email printed to stdout instead of actually sending them.
