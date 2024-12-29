# Beacon

![Beacon](imgs/Beacon-wide-bg.webp)

**Track health of your projects.**

Simple heartbeat and website monitor.

The goal of this app is to help you track the health of your projects (= everything is still running as it should). You can monitor:
- websites
- periodic jobs (cron or other)
- anything that can send HTTP POST request
- anything that can receive HTTP GET request

There are two main ways to use Beacon:
- Beacon checks your website or server for expected response = "web service".
- You sends health information (heartbeats) to Beacon = "heartbeat service". 

Beacon aims to be flexible and lets you choose how to use it.

## 🚀 Quickstart

Beacon can be easily installed anywhere [Go](https://go.dev/) is available.

```sh
# install as executable
go install github.com/davidmasek/beacon@latest

# start server
beacon start --config config.sample.yaml
```

You can always check current status of your services on the web GUI, by default on [http://localhost:8088](localhost:8088).
If you have SMTP configured, you will receive periodic reports via email.


To monitor your project, send heartbeats periodically from your application:
```sh
# use HTTP API (curl as an example, use anything you want)
curl -X POST http://localhost:8088/services/my-service-name/beat
```

To monitor your websites specify them in config file and Beacon will check them
periodically.

Beacon will automatically send you reports about your services. You can also
retrieve current status if needed.
```sh
# use HTTP API (curl as an example, use anything you want)
curl http://localhost:8088/services/my-service-name/status

# generate report for all your services
beacon report
# generate report and send it via email
# (see below for email configuration)
beacon report --send-mail
```

### Docker

Beacon is available as a Docker container. `compose.yaml` is provided for convenience. Simply start it with:
```sh
# for persistent database you need to mount it inside `compose.yaml`
docker compose up
```

For production usage you should mount your config file instead of `config.sample.yaml`.

You can also use docker directly without compose:
```sh
docker build -t beacon .
# add `-v $(pwd)/beacon.db:/root/beacon.db` if you want persist database
docker run --rm -p 8080:8080 -v $(pwd)/config.sample.yaml:/root/beacon.yaml beacon start
```

## Configuration

Configuration can be provided with CLI flags, environment variables and a config file. See [config.sample.yaml](config.sample.yaml) for example of a config file. By default the config file will be named `beacon.yaml` and located in your home directory (`~/beacon.yaml`). You can specify config file path with the `--config` CLI flag available for all commands.

If no config file is found in the default location (and you don't specify it yourself) then a default config file will be created. You can modify
the config file later as needed.

CLI flags take precedence over environment variables, which take precedence over config file. Environment variables should start with prefix `BEACON_` and use underscores for hierarchy. For example, to overwrite value for `smtp_port` under `email` section you would set `BEACON_SMTP_PORT` env variable. Anything specified inside config file can be overwritten using env variables.

### Service configuration

You can specify services in the `services` section. Each entry should be a unique service name, optionally with more configuration.

The minimal configuration for web services is just the `url` Beacon should check. The following example configures a service named `beacon-github`, that should be checked at the given url.

```yaml
services:
  beacon-github:
    url: "https://github.com/davidmasek/beacon"
```

If you're gonna provide service health info by sending heartbeats, no extra configuration is needed and you can specify only the service name as in the following example. Note that the line still ends with colon `:`, because it is a mapping of a name to a (in this case empty) configuration.

```yaml
services:
  truly-minimal-config:
```

The option `timeout` determines how long to consider a service healthy after a successful health check. It defaults to `24h` and needs to be specified with the unit included (`6h`, `24h`, `48h`, ...). For example, if a service has a timeout of 24 hours, it will be considered failed if it does not receive heartbeat for 24 hours.

`timeout` does not override health checks. For example if your website responds with unexpected status code (e.g. 404, 5xx, depending on settings) it will be immediately considered failed even if the `timeout` period did not pass yet.

The following fields are currently relevant only to web services:
- `url` - url to be periodically checked to determine service health
- `enabled` - set to `false` to disable automatically checking the website url
- `status` - HTTP status codes considered to be success (service healthy), defaults to 200
- `content` - content expected in the body of the response. Defaults to no checks. If multiple values are specified all need to be present.

### Email configuration

To receive emails from Beacon you need to provide SMTP server configuration. SMTP is a standard for sending email. You can find many SMTP providers online, both paid and free.

You can use Beacon without configuring email and use the web GUI, CLI or API to check status of your services.

The section `email` has the following fields:
- `smtp_server` - address of your SMTP server
- `smtp_port`
- `smtp_username`
- `smtp_password`
- `send_to` - your address where mail should be sent to
- `sender` - email address marked as sender of the emails
- `prefix` - any string, will be placed at start of the subject of each email. Useful to quickly differentiate different environments (I use it to separate dev/staging/prod).

## API

Beacon provides HTTP API to send and retrieve information about services.

Available endpoints:

|        Endpoint              |          Description          |
|------------------------------|-------------------------------|
| POST `/services/<id>/beat`   |  send heartbeat for service   | 
| GET  `/services/<id>/status` |  get latest health check      |


Examples:
```sh
❯ curl -X POST http://localhost:8088/services/sly-fox/beat
sly-fox @ 2024-12-29T18:55:11Z
❯ curl -X GET http://localhost:8088/services/sly-fox/status
sly-fox @ 2024-12-29T18:55:11Z
```

## Database

Beacon uses `beacon.db` file inside your home directory to store it's data. You can specify different path using the `BEACON_DB` env variable.

## 🌐 Website flow

You have a website. You point Beacon to it. Beacon continually checks that it is online. If it's not running you get a notification. 

🎉 That's it. You can now relax with the knowledge that everything is working.

🔍 You can optionally receive periodic reports to be really sure all is well.

🔍 This flow can also be used for other things than classic websites. Anything that can respond to HTTP GET requests goes.

## ❤️ Heartbeat flow

You have an application. You periodically notify Beacon that everything is good. If Beacon does not hear from the application for too long something is wrong and you get notified. 

🎉 That's it. You can now relax with the knowledge that everything is working.

🔍 You can optionally receive periodic reports to be really sure all is well.

🔍 This flow can be used for anything that can send HTTP POST requests.

## 🌟  Beacon status

Beacon is just starting and some things may be under construction 🚧, or not even that (= some features currently available only in your imagination 💭).

If you want to use Beacon you currently have to run, host and potentially (gasp) debug it yourself (although I do offer help if needed). A publicly available running instance is planned, so you don't have to do it all.

Development notes and more detailed status is available in [README-dev](README-dev.md).


## ⚙️ Build

```sh
# build a single binary called `beacon`
go build
```

## 🔬  Test

The Go tests are intended as the main testing component. They should run fast so you can
iterate quickly during development. They should cover main functionality (server, CLI, reporting, ...), including
reasonable fail scenarios (incorrect requests, config, ...).

Run Go tests:
```sh
go test ./...
# verbose mode
go test -v ./...
# disable cache
go test -count=1 ./...
```

Additionally there is a test script for Docker. This script has two goals. First, it tests
that the Docker image can be build and starts correctly, ensuring people can use Beacon via Docker.
Second, it runs Beacon in a clean environment, which helps catch problems that might be hidden
during local development.

The `test_docker.sh` script also tests sending email. For this, valid SMTP configuration is required.
See the script source if you are interested in running it.

Run testing script for Docker:
```sh
./test_docker.sh
```
