cmd ?= rest-server
args ?=
profile ?=
compose_args ?=

# Comma-separated Compose profiles; empty means development infrastructure only.
compose: export COMPOSE_PROFILES = $(profile)
compose: export TVTN_REST_ARGS = $(args)

.DEFAULT_GOAL := test

.PHONY: test run compose decompose

test:
	go test ./...

run:
	go run ./cmd/$(cmd) $(args)

compose:
	docker compose up -d --build $(compose_args)

# Deletes local database data and stack images. Build cache is shared across projects.
decompose:
	docker compose --profile '*' down --volumes --rmi all --remove-orphans
	docker builder prune --all --force
