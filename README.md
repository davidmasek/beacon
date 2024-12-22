# Beacon

![Beacon](imgs/Beacon-wide-bg.webp)

**Track health of your projects.**

Simple heartbeat and website monitor.

The goal of this app is to help you track the health of your projects (= everything is still running as it should). You can monitor:
- websites
- periodic jobs (cron or other)
- anything that can send HTTP POST request
- anything that can receive HTTP GET request

There are two main ways to use Beacon: website flow 🌐 and  heartbeat flow ❤️.

## 🚀 Quickstart

Beacon can be easily installed anywhere [Go](https://go.dev/) is available.

```sh
# install as executable
go install github.com/davidmasek/beacon@latest

# start server
beacon start --config config.sample.yaml
```

You can always check current status of your services on the web GUI, by default on [http://localhost:8089](localhost:8089).
If you have SMTP configured, you will receive periodic reports via email.


To monitor your project, send heartbeats periodically from your application:
```sh
# using Beacon CLI
beacon heartbeat my-service-name

# using HTTP API
# curl as an example, use anything you want
curl -X POST http://localhost:8088/beat/my-service-name
```

You can check status of your service(s), useful for programmatic access:
```sh
# using Beacon CLI
beacon status my-service-name

# using HTTP API
# curl as an example, use anything you want
curl http://localhost:8088/status/my-service-name

# generate report for all your services
beacon report
# generate report and send it via email
# (requires your SMTP server configuration)
beacon report --send-mail
```

You can also check health of websites:
```sh
# usage: beacon check <service-id> <url>
beacon check my-github https://github.com/davidmasek/beacon
beacon status my-github
```

### Docker

Beacon is available as a Docker container. `compose.yaml` is provided for convenience. Simply start it with:
```sh
docker compose up
```

For production usage you should mount your config file instead of `config.sample.yaml`.

You can also use docker directly without compose:
```sh
docker build -t beacon .
docker run --rm -p 8080:8080 -p 8089:8089 -v $(pwd)/config.sample.yaml:/root/beacon.yaml:ro beacon start
```

## Configuration

Configuration can be provided with CLI flags, environment variables and config files.

Details: WIP

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
