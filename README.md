# Beacon

![Beacon](imgs/Beacon-wide-bg.webp)

**Monitor your websites and periodic jobs with ease.**  

Beacon tracks the health of your websites, servers, and applications, so you know that everything runs as it should.

You can use Beacon to:
- Automatically check website availability and content.
- Receive notifications when services fail to send updates (heartbeats).
- Get periodic health reports via email or web GUI.

There are two main ways to use Beacon:
- Beacon checks your website or server for expected response = "web service".
- You sends health information (heartbeats) to Beacon = "heartbeat service". 

Beacon aims to be flexible and lets you choose how to use it.

## 🚀 Quickstart

Beacon can be easily installed anywhere [Go](https://go.dev/) is available.

1. **Install Beacon**
```sh
go install github.com/davidmasek/beacon@latest
```
2. **Start the Server**
```sh
# start server
beacon start --config config.sample.yaml
```
3. **Monitor Your Services**
- Website Monitoring: Use the configuration file to specify URLs for automatic periodic checks.
- Heartbeat Monitoring: Send periodic health updates (heartbeats) from your applications.
4. **Check the Status**
- Access the web GUI at http://localhost:8088
- Receive email notifications

### Docker

Beacon is also available as a Docker container. [`compose.yaml`](./compose.yaml) provides an example of using it with Docker Compose. The setup is runnable as-is and is intended for testing and development. For production I would recommend using images available from [Docker Hub](https://hub.docker.com/r/davidmasek42/beacon) and mounting database directory (`/app/db/`) for persistent storage.

```sh
docker compose up
```

## 🔧 Configuration

Beacon uses a configuration file to define monitored services and email settings for notifications. The default location is `~/beacon.yaml`, but you can specify a custom location using the `--config` CLI flag. If no configuration file is found, Beacon will create a default one.

### Examples

**Websites.** Verify that your homepage is accessible.
```yaml
services:
  my-homepage:
    url: "https://example.com/"
```

**APIs.** Verify that your API is accessible and returning expected content.
```yaml
services:
  my-api:
    url: "https://api.example.com/health"
    status: [200]
    content: ["healthy"]
```

**Periodic Jobs.** Ensure your cron jobs or other recurring tasks are running by sending a heartbeat after each run.
```yaml
services:
  nightly-backup:
```

### Service configuration

Services are defined in the services section of the config file. Each service must have a unique name.

**Web services**: Specify a url to periodically check the service.

```yaml
services:
  beacon-github:
    url: "https://github.com/davidmasek/beacon"
```

**Heartbeat services**: Specify only the service name.

Note that the line still ends with colon `:`, to ensure it is valid YAML file.

```yaml
services:
  my-app:
```

#### Additional Options

You can customize service behavior with optional fields.

| Field | Description | Default |
|-----------|-----------|-----------|
| `timeout` | Time (with units, e.g., `24h`, `48h`) after which a service is considered unhealthy.         | `24h`      |
| `enabled` | Set to `false` to temporarily disable monitoring for this service.                          | `true`     |
| `status`  | HTTP status codes that indicate the service is healthy.                                     | `200`      |
| `content` | Expected content in the response body (all values specified must be present).               | No checks  |


The option `timeout` determines how long to consider a service healthy after a successful health check. It defaults to `24h` and needs to be specified with the unit included (`6h`, `24h`, `48h`, ...). For example, if a service has a timeout of 24 hours, it will be considered failed if it does not receive heartbeat for 24 hours.

`timeout` does not override health checks. For example if your website responds with unexpected status code (e.g. 404, 5xx, depending on settings) it will be immediately considered failed even if the `timeout` period did not pass yet.

### Email configuration

Email notifications are optional but recommended for receiving health reports. Configure the email section in the config file with your SMTP server details. You can find many SMTP providers online, both paid and free.


| Field          | Description                                         | Required | Example                         |
|----------------|-----------------------------------------------------|----------|---------------------------------|
| `smtp_server`  | Address of the SMTP server                          | Yes      | `smtp.gmail.com`               |
| `smtp_port`    | SMTP port                                           | Yes      | `587`                          |
| `smtp_username`| SMTP server username                                | Yes      | `your-email@gmail.com`         |
| `smtp_password`| SMTP server password                                | Yes      | `your-password`                |
| `send_to`      | Recipient email address                             | Yes      | `your-email@gmail.com`         |
| `sender`       | Email address used as the sender                   | Yes      | `beacon@example.com`           |
| `prefix`       | String prepended to email subject for easy filtering | No       | `[Production]`                 |

Example:
```yaml
email:
  smtp_server: "smtp.gmail.com"
  smtp_port: 587
  smtp_username: "your-email@gmail.com"
  smtp_password: "your-password"
  send_to: "your-email@gmail.com"
  sender: "beacon@example.com"
  prefix: "[Production]"
```

You can use **environment variables** instead. For example:
```sh
export BEACON_EMAIL_SMTP_PASSWORD="your-password"
```

For password, you can instead provide a file containing the password.
```sh
export BEACON_EMAIL_SMTP_PASSWORD_FILE="/path/to/password-file"
```

### Other configuration

| Field          | Description                                          | Example                         |
|----------------|--------------------------------------------------|---------------------------------|
| `timezone`  | Timezone to use. Uses IANA Time Zone database, see https://pkg.go.dev/time#LoadLocation for details.                           | `Australia/Sydney`               |
| `report_time`    | Hour of the day after which to send reports. For example, `17` will be understood as 5pm, and reports will be send after 5pm as determined by `timezone`.                                             | `17`                          |


### Configuration sources

Beacon supports multiple configuration sources, with the following priority (highest to lowest):

1. CLI Flags (e.g., `--config`)
2. Environment Variables (e.g., `BEACON_SMTP_PASSWORD`)
3. Configuration File (default: `~/beacon.yaml`)

Environment variables use the prefix BEACON_ and replace dots (.) with underscores (_). For example, to override the smtp_port field in the email section:
```sh
export BEACON_EMAIL_SMTP_PORT=465
```

## API

Beacon provides an HTTP API to interact with your monitored services. You can use the API to send heartbeats, retrieve service statuses, and integrate Beacon into your workflows.

### Available endpoints

| Endpoint     | Method     | Description    |
|-----------|----------|---|
| `/services/<id>/beat`  | POST  |  Send a heartbeat for a specific service.   | 
| `/services/<id>/status` | GET |  Retrieve the latest health status of a service. |


### Examples

Send a Heartbeat:
```sh
curl -X POST http://localhost:8088/services/my-service-name/beat
```
Response:
```json
{
  "service_id": "my-service-name",
  "timestamp": "2025-01-11T17:20:09Z"
}
```

Get Service Status:
```sh
curl -X GET http://localhost:8088/services/my-service-name/status
```
Response:
```json
{
  "service_id": "my-service-name",
  "timestamp": "2025-01-11T17:20:09Z"
}
```


## Checking service status

Beacon will automatically send you reports about your services. You can always check current status of your services on the web GUI, by default on [http://localhost:8088](http://localhost:8088).

You can also check status of a specific service with the HTTP API or generate the full report via CLI. This enables you to build your own logic and reporting on top of Beacon as needed.

```sh
# view current status in your browser (default address)
http://localhost:8088

# use HTTP API (curl as an example, use anything you want)
curl http://localhost:8088/services/my-service-name/status

# generate report for all your services
beacon report
# generate report and send it via email
# (see below for email configuration)
beacon report --send-mail
```

## Database

Beacon uses `beacon.db` file inside your home directory to store it's data. 

You can specify different path using the `BEACON_DB` env variable.

If using Docker, mount the database file to persist data between container restarts.

## 🌟  Beacon status

Beacon is now stable. I use it for my personal projects. I will avoid breaking changes if possible, but backwards compatibility is not guaranteed till we reach version 1.0. 

Beacon is easily deployable as Docker container or installed as Go application.

Development notes and more detailed status is available in [README-dev](README-dev.md).

## ⚙️ Build

```sh
# build a single binary called `beacon`
go build
```

## 🔬  Test

The Go tests are intended as the main testing component. They should run fast so you can iterate quickly during development. They should cover main functionality (server, CLI, reporting, ...), including reasonable fail scenarios (incorrect requests, config, ...).

Run Go tests:
```sh
go test ./...
# verbose mode
go test -v ./...
# disable cache
go test -count=1 ./...
# with code coverage
go test ./... -coverprofile cover.out
```

Additionally there is a script for integration testing using Docker. This script has two goals. First, it tests
that the Docker image can be build and starts correctly, ensuring people can use Beacon via Docker.
Second, it runs Beacon in a clean environment, which helps catch problems that might be hidden
during local development.

Run testing script for Docker:
```sh
# tested with Python 3.10, any recent Python should work
python -m venv .venv
source .venv/bin/activate
python -m pip install -r requirements.dev.txt
# the script:
# 1. stops currently running containers with `docker compose down`
# 2. rebuilds the containers
# 3. runs tests
# 4. keeps containers running (for inspection, if needed)
python test_docker.py
```
