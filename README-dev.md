# Beacon - Dev Notes

This documents describes some internals, implementation details, and tries it's best to keep track of current feature list.

You should start with the main [README](README.md).

## 🚧 Feature list:
- 🟢 heartbeat listener
  - 🟢 HTTP server
  - 🟢 persistence
  - 🟡 stable API
    - needs finalization on "heartbeat"-only endpoints
    - later: needs endpoints for HealthCheck
- 🟡 web GUI
  - currently displays the main information
  - ~~should also support management~~
    - management will currently be supported only by CLI
  - support auth
    - blocked by *user management*
- 🟡 website monitor
  - needs more testing
  - needs periodic run solution
- 🟡 heartbeat/website management
  - currently hardcoded, needs more dynamic approach
  - needs refactor
  - TODO: 
- 🔴 user management
  - would enable multi-user server
  - would enable public server
- 🔴 friendly app configuration
  - need to decide on config form
  - needs documentation for config
  - nice-to-have: Dockerization
- 🔴 notifications
  - currently needs updates after refactors of other parts
- 🟡 dev workflow
  - 🟢 basic github setup
  - 🟢 CI for building/testing 
  - need more time to verify / refine
- 🟡 testing
  - 🟢 unit tests for storage
  - 🟡 unit tests for CLI (TODO: test)
  - 🟡 end-to-end "Go" test
    - done for heartbeat
    - want for web
    - maybe for reports in the future
  - 🔴 end-to-end CLI test
    - with Docker



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
