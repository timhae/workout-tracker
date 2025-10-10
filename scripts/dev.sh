#!/usr/bin/env bash

while true; do
    go run . &
    inotifywait -r -e modify --exclude '^./static/images/' .
    killall workout-tracker
done
