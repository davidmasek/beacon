#!/bin/bash
# Test building and starting Docker container.
# Test basic functionality.
# Requires working SMTP server.
set -e  # Exit immediately on any error
set -o pipefail  # Fail if any part of a pipeline fails

echo "Starting Beacon test."
echo "---------------------"

echo "Building Docker image..."
echo "---------------------"
DOCKER_BUILDKIT=1 docker build -t beacon-test .

# try to read env file, but ignore it if it does not exist
# in CI the env will be set without this file
source ~/beacon.github.env || echo "Ignore missing file in CI"

echo "Running test..."
echo "---------------------"
docker run --rm \
 --entrypoint bash \
 -e BEACON_EMAIL_SMTP_SERVER=${BEACON_EMAIL_SMTP_SERVER} \
 -e BEACON_EMAIL_SMTP_PORT=${BEACON_EMAIL_SMTP_PORT} \
 -e BEACON_EMAIL_SMTP_USERNAME=${BEACON_EMAIL_SMTP_USERNAME} \
 -e BEACON_EMAIL_SMTP_PASSWORD=${BEACON_EMAIL_SMTP_PASSWORD} \
 -e BEACON_EMAIL_SEND_TO=${BEACON_EMAIL_SEND_TO} \
 -e BEACON_EMAIL_SENDER=${BEACON_EMAIL_SENDER} \
 -e BEACON_EMAIL_PREFIX='[staging]' \
 beacon-test -c '
set -e
/app/beacon start &
echo "executing POST"
curl --retry 3 --retry-connrefused -sS -X POST http://localhost:8088/services/sly-fox/beat
echo "executing GET"
curl -sS -X GET http://localhost:8088/services/sly-fox/status
/app/beacon report --send-mail
'

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
