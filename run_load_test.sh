#!bin/bash

set -e
start_time=$(date -Iseconds)

echo "starting docker..."
docker compose -f docker-compose.test.yaml up -d



echo "waiting for avito-shop to start up"
timeout=60
interval=2
elapsed=0

while true; do
    if docker-compose logs --since="$start_time" avito-shop 2>&1 | grep -q "start listening"; then
        echo "avito-shop container is ready"
        break
    fi

    sleep $interval
    elapsed=$((elapsed+interval))

    if [ "$elapsed" -ge "$timeout" ]; then
        echo "wait timeout, stopping" >$2
        exit -1
    fi
done

echo "starting load tests"
k6 run --summary-export=./tests/load_test/result.json ./tests/load_test/load_test.js

echo "stopping docker"
docker-compose down
