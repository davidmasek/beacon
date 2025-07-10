# Beacon - Dev Notes

This documents describes some internals, implementation details, and tries it's best to keep track of current feature list.

You should start with the main [README](README.md).

Beware that this file may be out of date.

## Dev utilities

```sh
# non-structured logs
export BEACON_LOGS="dev"
export LOG_LEVEL="debug"
# hot-reload server
air start
```

## 🚧 Feature list:
- 🟢 heartbeat listener
  - 🟢 HTTP server
  - 🟢 persistence
  - 🟢 stable API
    - 🟢 needs finalization on "heartbeat"-only endpoints
    - 🟢 stabilize response - use JSON
    - 🟢 `/services/<id>/action` URL structure
  - 🟢 token auth
  - 🟢 (ignore unknown / require auth) if enabled
- 🟢 web GUI
  - 🟢 display the main information
    - management supported by a config file
  - 🟡 support auth
  - 🟢 unify ports - run on same port as HB listener
- 🟢 website monitor (periodic website checking)
  - 🟢 basic version done
  - 🟢 decoupled "web scraping" and reporting
- 🟢 notifications
  - 🟢 email reporting
  - 🟢 periodical monitoring
- 🟢 reports
  - 🟢 email reporting
- 🟢 heartbeat/website management
  - yellow - works, but needs some final touches
  - 🟢 specified in config
  - 🟡 some support for "manual" services without config - for heartbeats only
    - up to debate if these should be kept
- 🟢 friendly app configuration / documentation
  - 🟢 relative file paths handled
  - 🟢 config file refactor
    - 🟢 config file should required, but provided by default
  - 🟢 config format done
  - 🟢 main documentation done
  - 🟢 docker + dockerhub
  - 🟢 version info available
  - 🟢 nginx integration example
- 🟡 dev workflow
  - 🟢 basic github setup
  - 🟢 CI for building/testing 
  - need more time to verify / refine
  - 🟡 stabilize DB + versioning/migrations
- 🟡 testing
  - 🟢 unit tests for storage
  - 🟢 unit tests for CLI
  - 🟢 do not rely on external websites for unit testing
  - 🟢 code coverage 
    - manual checks, sufficient for now
  - 🟡 end-to-end "Go" test
    - done for heartbeat
    - want for web
    - maybe for reports in the future
  - 🟡 end-to-end Docker test
    - should cover also report content
- TODO: db cleanup
  - remove old record?
  - delete services option?

## 🛠️ Implementation

Some design choices:
- storage:
    - single type called HealthCheck for storing data in DB
    - need for different fields will be accommodated by using Metadata field (JSON)
    - for creating new data there is HealthCheckInput - currently same as HealthCheck without ID, in future possibly different
- naming conventions:
    - ID will be lowercased when used in variable name - FooId - to follow CamelCaseNaming
- modules:
    - storage (DB) is the base, handles persistence
    - monitors interact with the outside world and store health checks to DB
    - handlers
      - take data from DB and do something with it
      - display / generate reports
      - send notifications
    - cmd
      - entrypoints
      - should be simple, only wrap existing functionality
    - conf
      - store/load configuration
      - name chosen to prevent naming variables `config` (not super happy about naming here)


## Profiling

Kept here for possible future reference.

```sh
# cpu only
go test -cpuprofile=cpu.out ./scheduler
go tool pprof -http=:8080 ./scheduler.test cpu.out
# including blocking calls
go test -blockprofile=cpu.out ./scheduler
go tool pprof -http=:8080 ./scheduler.test block.out
# with trace, open "Goroutines" on the webpage
go test -trace=trace.out ./scheduler
go tool trace trace.out
```
