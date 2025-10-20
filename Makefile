.PHONY: test db dev

test:
	go test -v -shuffle=on -cover -coverprofile=cover.out
	go tool cover -html cover.out

db:
	docker volume create workout-tracker
	docker run --rm -it \
		-p 5432:5432 \
		-v workout-tracker:/var/lib/postgresql/data \
		-e POSTGRES_PASSWORD=postgres \
		postgres:17.6

dev:
	wgo run -xdir fixtures -file .html -file .css .
