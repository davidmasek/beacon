#!/usr/bin/env python3
import os
import subprocess
import time
import requests
import sys

DOCKER_COMPOSE_FILE = "compose.yaml"
BEACON_PORT = 8088
SMTP4DEV_PORT = 5080

def run_compose_cmd(*args):
    """Helper to run a docker compose command."""
    cmd = ["docker", "compose", "-f", DOCKER_COMPOSE_FILE] + list(args)
    print(f"> Running: {' '.join(cmd)}")
    subprocess.check_call(cmd)


def _post(url, headers=None):
    print(f"POST {url}")
    return requests.post(url, headers=headers)


def _get(url, headers=None):
    print(f"GET {url}")
    return requests.get(url, headers=headers)


def main():
    try:
        # print extra info about build
        os.environ.setdefault("BUILDKIT_PROGRESS", "plain")
        print("Starting Beacon test.")
        print("---------------------")
        run_compose_cmd("down")
        run_compose_cmd("up", "-d", "--build")

        print("Test API...")
        print("---------------------")

        for _ in range(10):
            try:
                resp = requests.get(f"http://localhost:{BEACON_PORT}")
                resp.raise_for_status()
                print("Beacon Up")
                break
            except requests.RequestException as ex:
                print("Waiting for Beacon to be ready...", ex)
                time.sleep(1)
        else:
            print("Beacon did not become ready in time.")
            sys.exit(1)

        # Basic flow
        post_url = f"http://localhost:{BEACON_PORT}/services/sly-fox/beat"
        resp = _post(post_url)
        resp.raise_for_status()
        print("POST succeeded!")

        get_url = f"http://localhost:{BEACON_PORT}/services/sly-fox/status"
        resp = _get(get_url)
        resp.raise_for_status()
        print("GET succeeded! Response:")
        print(resp.text)

        # Missing auth
        service_id = "heartbeat-with-auth"
        post_url = f"http://localhost:{BEACON_PORT}/services/{service_id}/beat"
        resp = _post(post_url)
        if resp.status_code != 401:
            raise ValueError(f"Expected request to fail, {post_url=}, {resp=}")
        print("POST failed as expected!")

        get_url = f"http://localhost:{BEACON_PORT}/services/{service_id}/status"
        resp = _get(get_url)
        if resp.status_code != 401:
            raise ValueError(f"Expected request to fail, {get_url=}, {resp=}")
        print("GET failed as expected! Response:")
        print(resp.text)

        # With auth
        headers = {
            "Authorization": "Bearer xasx@xJIXSA29jdfdnfNnuf$lLp"
        }
        post_url = f"http://localhost:{BEACON_PORT}/services/{service_id}/beat"
        resp = _post(post_url, headers=headers)
        resp.raise_for_status()
        print("POST succeeded!")

        get_url = f"http://localhost:{BEACON_PORT}/services/{service_id}/status"
        resp = _get(get_url, headers=headers)
        resp.raise_for_status()
        print("GET succeeded! Response:")
        print(resp.text)

        print("Test Sending report...")
        print("---------------------")

        # Wait for smtp4dev to be up
        smtp_health_url = f"http://localhost:{SMTP4DEV_PORT}/api/Version"
        for _ in range(10):
            try:
                resp = requests.get(smtp_health_url)
                resp.raise_for_status()
                print("SMTP Up")
                break
            except requests.RequestException:
                print("Waiting for smtp4dev to be ready...")
                time.sleep(1)
        else:
            print("smtp4dev did not become ready in time.")
            sys.exit(1)

        # Count emails before
        email_count_before = get_email_count()
        print(f"Email count before: {email_count_before}")

        # Trigger the report in the container
        # "docker compose exec beacon /app/beacon report --config /app/beacon.yaml"
        run_compose_cmd("exec", "beacon", "/app/beacon", "report", "--config", "/app/beacon.yaml")

        # Count emails after
        email_count_after = get_email_count()
        print(f"Email count after: {email_count_after}")

        if email_count_before == email_count_after:
            print("Testing: FAIL - email not sent")
            print("-------------")
            sys.exit(1)
        else:
            print("Testing: SUCCESS")
            print("----------------")
            sys.exit(0)

    except subprocess.CalledProcessError as e:
        print(f"Command failed: {e}")
        sys.exit(1)
    except requests.RequestException as e:
        print(f"HTTP request failed: {e}")
        sys.exit(1)


def get_email_count():
    """
    Fetches the list of messages from smtp4dev and counts
    how many contain 'you@example.fake' in the message data.
    """
    smtp_messages_url = f"http://localhost:{SMTP4DEV_PORT}/api/Messages"
    resp = requests.get(smtp_messages_url)
    resp.raise_for_status()

    messages_json = resp.json()
    results = messages_json["results"]

    return len(results)


if __name__ == "__main__":
    main()
