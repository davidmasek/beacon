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

## Quickstart

```sh
# install as executable
go install github.com/davidmasek/beacon@latest

# TODO: config

# start server
beacon start
```

### Docker

```sh
# TODO: copy config (.sample.yaml -> .yaml ?)
# TODO: use for testing
docker compose up --build
```

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

## 🚀 Run

```sh
go install github.com/davidmasek/beacon@latest

beacon start
```

## ⚙️ Build

```sh
# build a single binary called `beacon`
go build
```

## 🔬  Test

```sh
# go tests
go test ./...
# verbose mode
go test -v ./...
# disable cache
go test -count=1 ./...
```
