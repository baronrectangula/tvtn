cmd ?= rest-server
args ?=

.PHONY: test run compose decompose

test:
	go test ./...

run:
	go run ./cmd/$(cmd) $(args)

compose:
	docker compose up -d --build --wait

# Deletes local database data and stack images. Build cache is shared across projects.
decompose:
	docker compose down --volumes --rmi all --remove-orphans
	docker builder prune --all --force
