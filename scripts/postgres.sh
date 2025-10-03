#!/usr/bin/env bash

docker run --rm -it \
    -p 5432:5432 \
    -v $PWD/docker:/var/lib/postgresql/data \
    -e PGDATA=/var/lib/postgresql/data/pgdata \
    -e POSTGRES_PASSWORD=postgres postgres:17.6
