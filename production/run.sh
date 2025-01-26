#!/bin/bash

while true; do
    ./main 2>&1 | tee -a "log.txt"
    echo "$(date) Restarting in 5 seconds..."
    sleep 5
done
