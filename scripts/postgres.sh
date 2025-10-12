#!/usr/bin/env bash

docker volume create pgdata || true
docker run --rm -it \
    -p 5432:5432 \
    -v pgdata:/var/lib/postgresql/data \
    -e POSTGRES_PASSWORD=postgres \
    postgres:17.6
