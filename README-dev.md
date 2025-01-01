# Beacon - Dev Notes

This documents describes some internals, implementation details, and tries it's best to keep track of current feature list.

You should start with the main [README](README.md).

## 🚧 Feature list:
- 🟢 heartbeat listener
  - 🟢 HTTP server
  - 🟢 persistence
  - 🟡 stable API
    - 🟢 needs finalization on "heartbeat"-only endpoints
    - 🔴 stabilize response - use JSON
    - 🟤 later: needs endpoints for HealthCheck
    - 🟢 go with `/services/<id>/action` structure
- 🟢 web GUI
  - 🟢 display the main information
  - ~~should also support management~~
    - management will currently be supported only by CLI
  - 🟡 support auth
  - 🟢 unify ports - run on same port as HB listener
- 🟡 website monitor
  - needs more testing
- 🟡 reports
  - 🟢 basic flow
  - 🔴 should take config file into account (currently only looks at DB)
- 🟡 heartbeat/website management
  - 🟢 specified in config
  - 🟡 some support for "manual" services without config - for heartbeats only
    - up to debate if these should be kept
  - 🟤 delete old/unused service
- 🟡 periodic website checking
  - 🟢 basic version done
  - 🟢 should decouple "web scraping" and reporting
  - requires more work to be nice
- 🟡 user management
  - 🟢 DB prepared
  - would enable multi-user server
  - would enable public server
  - 🔴 auth
- 🔴 friendly app configuration / documentation
  - 🔴 TODO: DB needs some documentation
  - 🟡 TODO: relative file paths need some handling
  - 🟡 some "test my config file" functionality would be nice
  - 🔴 config file refactor
    - 🔴 config file should be required, but provided by default (inside homedir?)
    - 🔴 some utilities could be provided to check it / update it
  - 🟢 config format done
  - 🟢 main documentation done
  - 🟡 needs some user-testing to make sure it makes sense
  - 🟤 later: swagger API docs?
- 🟢 notifications
  - 🟢 email reporting
  - 🟢 local HTML report
  - 🟢 periodical monitoring
- 🟡 dev workflow
  - 🟢 basic github setup
  - 🟢 CI for building/testing 
  - need more time to verify / refine
  - 🔴 stabilize DB + versioning/migrations
- 🟡 testing
  - 🟢 unit tests for storage
  - 🟡 unit tests for CLI (TODO: test)
  - 🟡 end-to-end "Go" test
    - done for heartbeat
    - want for web
    - maybe for reports in the future
  - 🟡 end-to-end CLI test
    - with Docker
  - 🟡 test quality
    - some refactors would be nice
    - it might be good to not rely on external websites for unit testing, but not sure if it's worth it
      - https://pkg.go.dev/testing#hdr-Main
      - or just setup/teardown where needed...
  - 




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