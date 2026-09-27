# 🎟️ surge

A concurrent ticket reservation engine in Go with Postgres and Redis.

## Status

Actively being worked on. This is a portfolio project intended to demonstrate thoughtful system design with a high-quality Go implementation.
This is not intended to be used. However, it can be tested in its current state, allowing for the creation of events, venues, sections, along
with venue and event seats in bulk.

## Local development

```sh
make setup      # copy .env.example to .env.local
make docker-up  # start Postgres and Redis
make db-up      # run migrations
make start      # run the server
```

## Testing

```sh
make test   # run tests
make check  # run CI checks
```
