.PHONY: test db dev

test:
	go test -v -cover -test.coverprofile cover.out
	go tool cover -html cover.out

db:
	docker volume create pgdata || true; docker run --rm -it \
		-p 5432:5432 \
		-v pgdata:/var/lib/postgresql/data \
		-e POSTGRES_PASSWORD=postgres \
		postgres:17.6

dev:
	wgo run -xdir fixtures -file .html -file .css .
