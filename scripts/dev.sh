#!/usr/bin/env bash

while true; do
    go run . &
    inotifywait -r -e modify .
    killall workout-tracker
done
