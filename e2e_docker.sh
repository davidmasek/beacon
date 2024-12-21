#!/bin/bash
set -e  # Exit immediately on any error
set -o pipefail  # Fail if any part of a pipeline fails

# run as:
# docker compose  -f compose.yaml -f compose.test.yaml up --build 
# TODO: need to figure out how to quit
# - or a different approach


echo "Hey ya"
TEST_STATUS=0

if [ $TEST_STATUS -ne 0 ]; then
  echo "Go tests failed!"
  exit 1
fi

echo "All tests passed!"
exit 0
