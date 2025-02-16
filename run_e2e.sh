#!/bin/bash

#
# E2E TESTS SHOULD BE RUNNED WITH 2 ENV FLAGS
# POSTGRES_TEST_DB_URL=...
# REDIS_TEST_DB_URL=...
# STATUS=testing
#

set -e

echo "starting docker..."
docker compose --file docker-compose.test.yaml up -d --build

start_time=$(date -Iseconds)


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
    elapsed=$((elapsed+inverval))

    if [ "$elapsed" -ge "$timeout" ]; then
        echo "wait timeout, stopping" >$2
        exit -1
    fi
done

echo "starting e2e tests"
go test -v ./tests/e2e/...

echo "stopping docker"
docker-compose down
