# Beacon - Dev Notes

This documents describes some internals, implementation details, and tries it's best to keep track of current feature list.

You should start with the main [README](README.md).

Beware that this file may be out of date.

## Dev utilities

```sh
# non-structured logs
export BEACON_LOGS="dev"
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
    - 🟤 (low) endpoints for HealthCheck
    - 🟢 go with `/services/<id>/action` structure
  - 🔴 TODO auth
- 🟢 web GUI
  - 🟢 display the main information
  - ~~should also support management~~
    - management supported by config files
  - 🟡 support auth
  - 🟢 unify ports - run on same port as HB listener
- 🟢 website monitor (periodic website checking)
  - 🟢 basic version done
  - 🟢 should decouple "web scraping" and reporting
- 🟢 notifications
  - 🟢 email reporting
  - 🟢 local HTML report
  - 🟢 periodical monitoring
- 🟡 reports
  - yellow to keep eye on UX
  - 🟢 basic flow
  - 🟢 should take config file into account (currently only looks at DB)
- 🟡 heartbeat/website management
  - yellow - works, but needs some final touches
  - 🟢 specified in config
  - 🟡 some support for "manual" services without config - for heartbeats only
    - up to debate if these should be kept
  - 🟤 delete old/unused service
- 🟡 friendly app configuration / documentation
  - 🟡 DB needs some documentation
  - 🟢 relative file paths handled
  - 🟢 config file refactor
    - 🟢 config file should be required, but provided by default (inside homedir?)
  - 🟢 config format done
  - 🟢 main documentation done
  - 🟢 docker + dockerhub
  - 🟢 version info available
  - 🟤 todo: nginx integration example?
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
- 🟤 user management
  - 🟢 DB prepared
  - would enable multi-user server
  - would enable public server
  - 🟤 auth
  - 🟤 actual usage of users
  - 🟤 registration / login


## Run with Live Reload

```sh
# see https://github.com/air-verse/air for details
air start
```


## 🛠️ Implementation

Some design choices:
- storage:
    - single type called HealthCheck for storing data in DB
    - need for different fields will be accommodated by using Metadata field, which is dynamic map
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
