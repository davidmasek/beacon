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

🚧 Feature list:
- 🟢 heartbeat listener
  - 🟢 HTTP server
  - 🟢 persistence
  - 🟡 stable API
- 🟡 web GUI
  - currently displays the main information
  - should also support management
  - related to the other "management" features
- 🟡 website monitor
  - needs more testing
  - needs periodic run solution
- 🟡 heartbeat/website management
  - currently hardcoded, needs more dynamic approach
  - needs refactor
- 🔴 user management
  - this is currently a blocker for stable API and public instance
- 🔴 friendly app configuration
  - this is an inconvenience for potential users
- 🔴 notifications
  - currently needs updates after refactors of other parts
- 🟡 dev workflow
  - 🟢 basic github setup
  - want CI for building/testing 
- 🟡 testing
  - 🟢 unit tests for storage
  - want at least one end-to-end test
  - want more automation, related to "dev workflow"


## 🚀 Run

```sh
# for server - with hot reload
air

# run monitors (i.e. check stuff)
go run ./cmd/monitor
# server - display UI + listen for heartbeats
go run ./cmd/server
```

## ⚙️ Build

```sh
go build -v -o bin ./...
```

## 🔬  Test

```sh
go test -v ./...
```

## 🛠️ Dev

Some design choices:
- storage:
    - single type called HealthCheck for storing data in DB
    - need for different fields will be accommodated by using Metadata field, which is dynamic map
    - for creating new data there is HealthCheckInput - currently same as HealthCheck without ID, in future possibly different
- naming conventions:
    - ID will be lowercased when used in variable name - FooId - to follow CamelCaseNaming
- dependency chain / architecture:
    - storage < monitor < handlers < status < cmd
    - storage (DB) is the base, handles persistence, should depend on nothing (nothing internal, can depend e.g. on SQLite)
    - monitors interact with the outside world and store health checks to DB
    - handlers / status
        - the idea was that "status" evaluates if something is OK or not and "handlers" then handle the result
        - a bit of a mess currently
        - should read DB and react in some way:
        - display / generate reports
        - send notifications
    - cmd
        - entrypoints
        - can depend on anything (apart from each other)
        - should be simple and high-level

