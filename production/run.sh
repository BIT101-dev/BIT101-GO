#!/bin/bash

# 创建日志目录（如果不存在）
LOG_DIR="./log"
mkdir -p "$LOG_DIR"

while true; do
    LOG_FILE="$LOG_DIR/$(date +"%Y%m%d").txt"
    ./main >> "$LOG_FILE" 2>&1
    echo "$(date) Restarting in 5 seconds..."
    sleep 5
done
