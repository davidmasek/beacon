#!/bin/bash
# Test building and starting Docker container.
# Test basic functionality.
# Requires working SMTP server.
set -e  # Exit immediately on any error
set -o pipefail  # Fail if any part of a pipeline fails

echo "Starting Beacon test."
echo "---------------------"

# DOCKER_BUILDKIT=1 docker compose build beacon
docker compose run --rm \
 -T \
 --entrypoint bash \
 -v $(pwd)/config.yaml:/root/beacon.yaml:ro \
 beacon <<EOF
set -e
/app/beacon start &
curl -sS -X POST http://localhost:8088/beat/sly-fox
curl -sS -X GET http://localhost:8088/status/sly-fox
/app/beacon report --send-mail
EOF
TEST_RESULT=$?
if [ "$TEST_RESULT" -eq "0" ]; then
    echo "Testing: SUCCESS"
    echo "----------------"
else
    echo "Testing: FAIL"
    echo "-------------"
fi

echo "Done."
exit $TEST_RESULT
